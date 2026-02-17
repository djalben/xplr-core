package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/notification"
	"github.com/shopspring/decimal"
)

// GetCardByID извлекает карту со всеми полями
func GetCardByID(id int) (models.Card, error) {
	if GlobalDB == nil {
		return models.Card{}, fmt.Errorf("database connection not initialized")
	}
	var card models.Card
	var teamID sql.NullInt64
	query := `
		SELECT id, user_id, provider_card_id, bin, last_4_digits, card_status, 
		       COALESCE(nickname, '') as nickname, COALESCE(service_slug, 'arbitrage'), daily_spend_limit, failed_auth_count,
		       COALESCE(card_type, 'VISA') as card_type,
		       COALESCE(auto_replenish_enabled, FALSE) as auto_replenish_enabled,
		       COALESCE(auto_replenish_threshold, 0) as auto_replenish_threshold,
		       COALESCE(auto_replenish_amount, 0) as auto_replenish_amount,
		       COALESCE(card_balance, 0) as card_balance,
		       team_id, created_at
		FROM cards WHERE id = $1
	`
	err := GlobalDB.QueryRow(query, id).Scan(
		&card.ID, &card.UserID, &card.ProviderCardID, &card.BIN, &card.Last4Digits,
		&card.CardStatus, &card.Nickname, &card.ServiceSlug, &card.DailySpendLimit, &card.FailedAuthCount,
		&card.CardType, &card.AutoReplenishEnabled, &card.AutoReplenishThreshold,
		&card.AutoReplenishAmount, &card.CardBalance, &teamID, &card.CreatedAt,
	)
	if teamID.Valid {
		teamIDVal := int(teamID.Int64)
		card.TeamID = &teamIDVal
	}
	return card, err
}

// GetUserCards извлекает все карты пользователя. По архитектуре XPLR баланс каждой карты
// виртуальный — в ответе card_balance = BalanceRub пользователя (как в Platipomiru).
func GetUserCards(userID int) ([]models.Card, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, provider_card_id, bin, last_4_digits, card_status, 
		       COALESCE(nickname, '') as nickname, COALESCE(service_slug, 'arbitrage'), daily_spend_limit, failed_auth_count,
		       COALESCE(card_type, 'VISA') as card_type,
		       COALESCE(auto_replenish_enabled, FALSE) as auto_replenish_enabled,
		       COALESCE(auto_replenish_threshold, 0) as auto_replenish_threshold,
		       COALESCE(auto_replenish_amount, 0) as auto_replenish_amount,
		       COALESCE(card_balance, 0) as card_balance,
		       team_id, created_at
		FROM cards 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`
	rows, err := GlobalDB.Query(query, userID)
	if err != nil {
		log.Printf("DB Error fetching cards for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to fetch cards")
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		var teamID sql.NullInt64
		err := rows.Scan(
			&card.ID,
			&card.UserID,
			&card.ProviderCardID,
			&card.BIN,
			&card.Last4Digits,
			&card.CardStatus,
			&card.Nickname,
			&card.ServiceSlug,
			&card.DailySpendLimit,
			&card.FailedAuthCount,
			&card.CardType,
			&card.AutoReplenishEnabled,
			&card.AutoReplenishThreshold,
			&card.AutoReplenishAmount,
			&card.CardBalance,
			&teamID,
			&card.CreatedAt,
		)
		if teamID.Valid {
			teamIDVal := int(teamID.Int64)
			card.TeamID = &teamIDVal
		}
		if err != nil {
			log.Printf("DB Error scanning card: %v", err)
			continue
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	// XPLR: виртуальный баланс — у каждой карты в списке отображаем общий BalanceRub пользователя
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	for i := range cards {
		cards[i].CardBalance = user.BalanceRub
	}

	return cards, nil
}

// ProcessCardPayment — Атомарное списание средств с баланса пользователя
// Использует транзакцию БД и SELECT ... FOR UPDATE для предотвращения race conditions
func ProcessCardPayment(userID int, cardID int, amount decimal.Decimal, fee decimal.Decimal, merchantName string, cardLast4 string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// 1. НАЧАЛО ТРАНЗАКЦИИ
	tx, err := GlobalDB.Begin()
	if err != nil {
		log.Printf("DB Error Begin Transaction: %v", err)
		return fmt.Errorf("не удалось начать транзакцию")
	}
	defer tx.Rollback() // Откатываем, если не будет Commit

	// 2. БЛОКИРОВКА СТРОКИ ПОЛЬЗОВАТЕЛЯ (SELECT ... FOR UPDATE)
	// XPLR: списание с balance_rub (основной баланс в рублях)
	var currentBalance decimal.Decimal
	err = tx.QueryRow(
		"SELECT COALESCE(balance_rub, 0) FROM users WHERE id = $1 FOR UPDATE",
		userID,
	).Scan(&currentBalance)

	if err != nil {
		log.Printf("DB Error Locking User Row: %v", err)
		return fmt.Errorf("не удалось заблокировать запись пользователя")
	}

	// 3. ПРОВЕРКА БАЛАНСА (двойная проверка для безопасности)
	if currentBalance.LessThan(amount) {
		log.Printf("CRITICAL: Insufficient balance after lock. User %d, Balance: %s, Amount: %s",
			userID, currentBalance.String(), amount.String())
		return fmt.Errorf("недостаточно средств на балансе")
	}

	// 4. СПИСАНИЕ СРЕДСТВ (XPLR: balance_rub)
	_, err = tx.Exec(
		"UPDATE users SET balance_rub = COALESCE(balance_rub, 0) - $1, balance = balance - $2 WHERE id = $3",
		amount, amount, userID,
	)
	if err != nil {
		log.Printf("DB Error Deducting Balance: %v", err)
		return fmt.Errorf("не удалось списать средства")
	}

	// 5. ЗАПИСЬ ТРАНЗАКЦИИ (с комиссией на основе Grade)
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, executed_at)
		 VALUES ($1, $2, $3, $4, 'CAPTURE', 'APPROVED', $5, $6)`,
		userID,
		cardID,
		amount,
		fee, // Комиссия на основе Grade пользователя
		fmt.Sprintf("Card payment: %s from ...%s", merchantName, cardLast4),
		time.Now(),
	)
	if err != nil {
		log.Printf("DB Error Recording Transaction: %v", err)
		return fmt.Errorf("не удалось записать транзакцию")
	}

	// 6. СБРОС СЧЕТЧИКА НЕУДАЧНЫХ АВТОРИЗАЦИЙ
	_, err = tx.Exec(
		"UPDATE cards SET failed_auth_count = 0 WHERE id = $1",
		cardID,
	)
	if err != nil {
		log.Printf("DB Error Resetting Failed Auth Count: %v", err)
		return fmt.Errorf("не удалось обновить счетчик карты")
	}

	// 7. КОММИТ ТРАНЗАКЦИИ
	if err := tx.Commit(); err != nil {
		log.Printf("DB Error Commit: %v", err)
		return fmt.Errorf("ошибка фиксации транзакции")
	}

	log.Printf("✅ Payment processed successfully. User %d, Amount: %s, Merchant: %s",
		userID, amount.String(), merchantName)

	// 8. Обновить Grade пользователя (в фоне, не блокируем ответ)
	// Импорт UpdateUserGrade из grade.go (он в том же пакете repository)
	go func() {
		if err := UpdateUserGrade(userID); err != nil {
			log.Printf("Warning: Failed to update user grade for user %d: %v", userID, err)
		}
	}()

	return nil
}

// IncrementFailedAuthCount увеличивает счетчик ошибок
func IncrementFailedAuthCount(cardID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec("UPDATE cards SET failed_auth_count = failed_auth_count + 1 WHERE id = $1", cardID)
	return err
}

// BlockCard блокирует карту (устанавливает статус BLOCKED)
func BlockCard(cardID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec("UPDATE cards SET card_status = 'BLOCKED' WHERE id = $1", cardID)
	if err != nil {
		log.Printf("DB Error Blocking Card %d: %v", cardID, err)
		return fmt.Errorf("не удалось заблокировать карту")
	}
	log.Printf("🔒 Card %d has been BLOCKED due to multiple failed attempts", cardID)
	return nil
}

// UpdateCardStatus sets card_status to "BLOCKED" or "ACTIVE" for a card owned by userID.
func UpdateCardStatus(cardID int, userID int, status string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	if status != "BLOCKED" && status != "ACTIVE" {
		return fmt.Errorf("invalid status: must be BLOCKED or ACTIVE")
	}
	res, err := GlobalDB.Exec(
		"UPDATE cards SET card_status = $1 WHERE id = $2 AND user_id = $3",
		status, cardID, userID,
	)
	if err != nil {
		log.Printf("DB Error UpdateCardStatus card %d: %v", cardID, err)
		return fmt.Errorf("failed to update card status")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("card not found or access denied")
	}
	log.Printf("✅ Card %d status updated to %s (user %d)", cardID, status, userID)
	return nil
}

// IssueCards — Mock-функция для выпуска виртуальных карт (без реального банка)
// Генерирует фейковые карты с BIN 4242 для тестирования фронтенда
func IssueCards(userID int, req models.MassIssueRequest) (interface{}, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	log.Printf("IssueCards: User %d requested %d cards", userID, req.Count)

	var results []models.CardIssueResult
	successCount := 0
	failedCount := 0

	for i := 0; i < req.Count; i++ {
		// Генерируем случайные последние 4 цифры карты
		last4 := fmt.Sprintf("%04d", (userID*1000+i)%10000)

		// Вставляем карту в БД
		var cardID int
		var createdAt time.Time

		cardType := req.CardType
		if cardType == "" {
			cardType = "VISA"
		}
		serviceSlug := req.ServiceSlug
		if serviceSlug == "" {
			serviceSlug = "arbitrage"
		}

		// Если указан team_id, проверяем доступ пользователя к команде
		if req.TeamID != nil && *req.TeamID > 0 {
			hasAccess, _, err := CheckTeamAccess(*req.TeamID, userID)
			if err != nil || !hasAccess {
				log.Printf("Access denied: User %d does not have access to team %d", userID, *req.TeamID)
				failedCount++
				results = append(results, models.CardIssueResult{
					Success:   false,
					Status:    "FAILED",
					CardLast4: last4,
					Nickname:  req.CardNickname,
					Message:   "Access denied to team",
				})
				continue
			}
		}

		err := GlobalDB.QueryRow(`
			INSERT INTO cards (user_id, provider_card_id, bin, last_4_digits, card_status, nickname, service_slug, daily_spend_limit, failed_auth_count, card_type, card_balance, team_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id, created_at
		`,
			userID,
			fmt.Sprintf("MOCK-%d-%s", userID, last4),
			"424242",
			last4,
			"ACTIVE",
			req.CardNickname,
			serviceSlug,
			req.DailyLimit,
			0,
			cardType,
			decimal.Zero,
			req.TeamID,
		).Scan(&cardID, &createdAt)

		if err != nil {
			log.Printf("Failed to insert card for user %d: %v", userID, err)
			failedCount++
			results = append(results, models.CardIssueResult{
				Success:   false,
				Status:    "FAILED",
				CardLast4: last4,
				Nickname:  req.CardNickname,
				Message:   fmt.Sprintf("Failed to issue card: %v", err),
			})
			continue
		}

		// Успешно создана карта
		successCount++
		results = append(results, models.CardIssueResult{
			Success:   true,
			Status:    "ACTIVE",
			CardLast4: last4,
			Nickname:  req.CardNickname,
			Message:   "Card issued successfully",
			Card: &models.Card{
				ID:              cardID,
				UserID:          userID,
				TeamID:          req.TeamID,
				BIN:             "424242",
				Last4Digits:     last4,
				CardStatus:      "ACTIVE",
				ServiceSlug:     serviceSlug,
				DailySpendLimit: req.DailyLimit,
				FailedAuthCount: 0,
				CardType:        cardType,
				CardBalance:     decimal.Zero,
				CreatedAt:       createdAt,
			},
		})
	}

	response := models.MassIssueResponse{
		Successful: successCount,
		Failed:     failedCount,
		Results:    results,
	}

	log.Printf("✅ Issued %d cards successfully, %d failed for user %d", successCount, failedCount, userID)

	// ОТПРАВКА TELEGRAM УВЕДОМЛЕНИЯ
	if successCount > 0 {
		user, err := GetUserByID(userID)
		if err == nil && user.TelegramChatID.Valid {
			message := fmt.Sprintf("💳 Cards Issued Successfully!\n\nCount: %d cards\nMerchant: %s\nDaily Limit: $%.2f per card",
				successCount,
				req.MerchantName,
				req.DailyLimit)
			notification.SendTelegramMessage(user.TelegramChatID.Int64, message)
		}
	}

	return response, nil
}

// UpdateCardAutoReplenishment - Обновить настройки автопополнения карты
func UpdateCardAutoReplenishment(cardID int, userID int, enabled bool, threshold decimal.Decimal, amount decimal.Decimal) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Проверяем, что карта принадлежит пользователю
	var cardUserID int
	err := GlobalDB.QueryRow("SELECT user_id FROM cards WHERE id = $1", cardID).Scan(&cardUserID)
	if err != nil {
		return fmt.Errorf("card not found")
	}
	if cardUserID != userID {
		return fmt.Errorf("access denied: card does not belong to user")
	}

	// Обновляем настройки автопополнения
	_, err = GlobalDB.Exec(
		`UPDATE cards 
		 SET auto_replenish_enabled = $1, 
		     auto_replenish_threshold = $2, 
		     auto_replenish_amount = $3 
		 WHERE id = $4 AND user_id = $5`,
		enabled, threshold, amount, cardID, userID,
	)
	if err != nil {
		log.Printf("DB Error updating auto-replenishment for card %d: %v", cardID, err)
		return fmt.Errorf("failed to update auto-replenishment settings")
	}

	log.Printf("✅ Auto-replenishment updated for card %d: enabled=%v, threshold=%s, amount=%s", 
		cardID, enabled, threshold.String(), amount.String())
	return nil
}

// GetCardsNeedingReplenishment - Получить карты, требующие пополнения
func GetCardsNeedingReplenishment() ([]models.Card, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, provider_card_id, bin, last_4_digits, card_status, 
		       COALESCE(nickname, '') as nickname, COALESCE(service_slug, 'arbitrage'), daily_spend_limit, failed_auth_count,
		       COALESCE(card_type, 'VISA') as card_type,
		       auto_replenish_enabled, auto_replenish_threshold, auto_replenish_amount,
		       COALESCE(card_balance, 0) as card_balance, team_id, created_at
		FROM cards 
		WHERE auto_replenish_enabled = TRUE 
		  AND card_status = 'ACTIVE'
		  AND COALESCE(card_balance, 0) <= auto_replenish_threshold
	`
	rows, err := GlobalDB.Query(query)
	if err != nil {
		log.Printf("DB Error fetching cards needing replenishment: %v", err)
		return nil, fmt.Errorf("failed to fetch cards needing replenishment")
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		var teamID sql.NullInt64
		err := rows.Scan(
			&card.ID, &card.UserID, &card.ProviderCardID, &card.BIN, &card.Last4Digits,
			&card.CardStatus, &card.Nickname, &card.ServiceSlug, &card.DailySpendLimit, &card.FailedAuthCount,
			&card.CardType, &card.AutoReplenishEnabled, &card.AutoReplenishThreshold,
			&card.AutoReplenishAmount, &card.CardBalance, &teamID, &card.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning card: %v", err)
			continue
		}
		if teamID.Valid {
			teamIDVal := int(teamID.Int64)
			card.TeamID = &teamIDVal
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// ReplenishCard - Пополнить карту (увеличить card_balance)
func ReplenishCard(cardID int, amount decimal.Decimal) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	_, err := GlobalDB.Exec(
		"UPDATE cards SET card_balance = card_balance + $1 WHERE id = $2",
		amount, cardID,
	)
	if err != nil {
		log.Printf("DB Error replenishing card %d: %v", cardID, err)
		return fmt.Errorf("failed to replenish card")
	}

	log.Printf("✅ Card %d replenished with amount %s", cardID, amount.String())
	return nil
}