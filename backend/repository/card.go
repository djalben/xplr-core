package repository

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/domain"
	"github.com/djalben/xplr-core/backend/notification"
	"github.com/djalben/xplr-core/backend/telegram"
	"github.com/shopspring/decimal"
)

// GetCardByID извлекает карту со всеми полями
func GetCardByID(id int) (domain.Card, error) {
	if GlobalDB == nil {
		return domain.Card{}, fmt.Errorf("database connection not initialized")
	}
	var card domain.Card
	var teamID sql.NullInt64
	query := `
		SELECT id, user_id, provider_card_id, bin, last_4_digits, card_status, 
		       COALESCE(nickname, '') as nickname, COALESCE(service_slug, 'arbitrage'), daily_spend_limit, COALESCE(spend_limit, 0) as spend_limit, failed_auth_count,
		       COALESCE(card_type, 'VISA') as card_type,
		       COALESCE(category, 'arbitrage') as category,
		       COALESCE(currency, 'USD') as currency,
		       COALESCE(auto_replenish_enabled, FALSE) as auto_replenish_enabled,
		       COALESCE(auto_replenish_threshold, 0) as auto_replenish_threshold,
		       COALESCE(auto_replenish_amount, 0) as auto_replenish_amount,
		       COALESCE(card_balance, 0) as card_balance,
		       team_id, created_at
		FROM cards WHERE id = $1
	`
	err := GlobalDB.QueryRow(query, id).Scan(
		&card.ID, &card.UserID, &card.ProviderCardID, &card.BIN, &card.Last4Digits,
		&card.CardStatus, &card.Nickname, &card.ServiceSlug, &card.DailySpendLimit, &card.SpendLimit, &card.FailedAuthCount,
		&card.CardType, &card.Category, &card.Currency, &card.AutoReplenishEnabled, &card.AutoReplenishThreshold,
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
func GetUserCards(userID int) ([]domain.Card, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, provider_card_id, bin, last_4_digits, card_status, 
		       COALESCE(nickname, '') as nickname, COALESCE(service_slug, 'arbitrage'), daily_spend_limit, COALESCE(spend_limit, 0) as spend_limit, failed_auth_count,
		       COALESCE(card_type, 'VISA') as card_type,
		       COALESCE(category, 'arbitrage') as category,
		       COALESCE(currency, 'USD') as currency,
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

	var cards []domain.Card
	for rows.Next() {
		var card domain.Card
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
			&card.SpendLimit,
			&card.FailedAuthCount,
			&card.CardType,
			&card.Category,
			&card.Currency,
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

	return cards, nil
}

// GetFirstActiveCard — returns the first ACTIVE card for a user (for store purchases).
// IMPORTANT: Travel cards (service_slug='travel') are EXCLUDED from store purchases.
// Only subscription/premium/arbitrage cards are allowed.
func GetFirstActiveCard(userID int) (*domain.Card, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	var card domain.Card
	var teamID sql.NullInt64
	err := GlobalDB.QueryRow(`
		SELECT id, user_id, provider_card_id, bin, last_4_digits, card_status,
		       COALESCE(nickname, '') as nickname, COALESCE(service_slug, 'arbitrage'),
		       daily_spend_limit, COALESCE(spend_limit, 0) as spend_limit, failed_auth_count,
		       COALESCE(card_type, 'VISA') as card_type,
		       COALESCE(category, 'arbitrage') as category,
		       COALESCE(currency, 'USD') as currency,
		       COALESCE(auto_replenish_enabled, FALSE) as auto_replenish_enabled,
		       COALESCE(auto_replenish_threshold, 0) as auto_replenish_threshold,
		       COALESCE(auto_replenish_amount, 0) as auto_replenish_amount,
		       COALESCE(card_balance, 0) as card_balance,
		       team_id, created_at
		FROM cards
		WHERE user_id = $1 AND card_status = 'ACTIVE'
		  AND COALESCE(service_slug, 'arbitrage') NOT IN ('travel')
		ORDER BY created_at ASC
		LIMIT 1
	`, userID).Scan(
		&card.ID, &card.UserID, &card.ProviderCardID, &card.BIN, &card.Last4Digits,
		&card.CardStatus, &card.Nickname, &card.ServiceSlug, &card.DailySpendLimit, &card.SpendLimit, &card.FailedAuthCount,
		&card.CardType, &card.Category, &card.Currency, &card.AutoReplenishEnabled, &card.AutoReplenishThreshold,
		&card.AutoReplenishAmount, &card.CardBalance, &teamID, &card.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // no active card — not an error
		}
		return nil, fmt.Errorf("failed to query active card: %w", err)
	}
	if teamID.Valid {
		v := int(teamID.Int64)
		card.TeamID = &v
	}
	return &card, nil
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

// BlockCard блокирует карту (устанавливает статус BLOCKED).
// Delegates to UpdateCardStatus which atomically refunds card_balance to wallet.
func BlockCard(cardID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	// Look up card owner so we can use the ACID UpdateCardStatus path
	var userID int
	err := GlobalDB.QueryRow("SELECT user_id FROM cards WHERE id = $1", cardID).Scan(&userID)
	if err != nil {
		log.Printf("DB Error looking up owner of card %d for block: %v", cardID, err)
		return fmt.Errorf("card not found")
	}
	if err := UpdateCardStatus(cardID, userID, "BLOCKED"); err != nil {
		log.Printf("DB Error blocking card %d via UpdateCardStatus: %v", cardID, err)
		return err
	}
	log.Printf("🔒 Card %d has been BLOCKED (ACID refund applied) for user %d", cardID, userID)
	return nil
}

// UpdateCardStatus sets card_status for a card owned by userID.
// Supported statuses: ACTIVE, BLOCKED, FROZEN, CLOSED.
// On CLOSED or BLOCKED: atomically refunds card_balance back to wallet (master_balance).
// Issue fee is NOT refunded — only remaining card balance.
func UpdateCardStatus(cardID int, userID int, status string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	validStatuses := map[string]bool{"ACTIVE": true, "BLOCKED": true, "FROZEN": true, "CLOSED": true}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: must be ACTIVE, BLOCKED, FROZEN or CLOSED")
	}

	tx, err := GlobalDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Lock card row and verify ownership
	var cardBalance decimal.Decimal
	var last4 string
	err = tx.QueryRow(
		`SELECT COALESCE(card_balance, 0), last_4_digits FROM cards WHERE id = $1 AND user_id = $2 FOR UPDATE`,
		cardID, userID,
	).Scan(&cardBalance, &last4)
	if err != nil {
		return fmt.Errorf("card not found or access denied")
	}

	// Refund card_balance to wallet on CLOSED or BLOCKED
	if (status == "CLOSED" || status == "BLOCKED") && cardBalance.GreaterThan(decimal.Zero) {
		// Credit master_balance
		_, err = tx.Exec(
			`UPDATE internal_balances SET master_balance = master_balance + $1, updated_at = NOW() WHERE user_id = $2`,
			cardBalance, userID,
		)
		if err != nil {
			log.Printf("DB Error refunding card %d balance to wallet: %v", cardID, err)
			return fmt.Errorf("failed to refund card balance to wallet")
		}

		// Zero out card_balance
		_, err = tx.Exec(
			`UPDATE cards SET card_balance = 0 WHERE id = $1`,
			cardID,
		)
		if err != nil {
			return fmt.Errorf("failed to zero card balance")
		}

		// Record CARD_REFUND transaction
		details := fmt.Sprintf("Возврат остатка $%s с карты •••• %s при %s",
			cardBalance.StringFixed(2), last4, map[string]string{"CLOSED": "закрытии", "BLOCKED": "блокировке"}[status])
		_, err = tx.Exec(
			`INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, executed_at)
			 VALUES ($1, $2, $3, 0, 'CARD_REFUND', 'APPROVED', $4, $5)`,
			userID, cardID, cardBalance, details, time.Now(),
		)
		if err != nil {
			log.Printf("DB Error recording CARD_REFUND for card %d: %v", cardID, err)
		}

		log.Printf("💰 Card %d: refunded $%s to wallet (user %d) on %s",
			cardID, cardBalance.StringFixed(2), userID, status)
	}

	// Update card status
	_, err = tx.Exec(
		`UPDATE cards SET card_status = $1 WHERE id = $2 AND user_id = $3`,
		status, cardID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update card status")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %v", err)
	}

	log.Printf("✅ Card %d status updated to %s (user %d)", cardID, status, userID)
	return nil
}

// UpdateCardSpendLimit updates the spend_limit for a card owned by userID.
func UpdateCardSpendLimit(cardID int, userID int, limit decimal.Decimal) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	res, err := GlobalDB.Exec(
		"UPDATE cards SET spend_limit = $1, daily_spend_limit = $1 WHERE id = $2 AND user_id = $3",
		limit, cardID, userID,
	)
	if err != nil {
		log.Printf("DB Error UpdateCardSpendLimit card %d: %v", cardID, err)
		return fmt.Errorf("failed to update spend limit")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("card not found or access denied")
	}
	log.Printf("✅ Card %d spend_limit updated to %s (user %d)", cardID, limit.String(), userID)

	// Telegram notification (graceful - won't fail if bot token not set)
	user, err := GetUserByID(userID)
	if err == nil && user.TelegramChatID.Valid {
		card, cerr := GetCardByID(cardID)
		last4 := "????"
		if cerr == nil {
			last4 = card.Last4Digits
		}
		notification.SendTelegramMessage(user.TelegramChatID.Int64,
			fmt.Sprintf("⚡ Лимит карты обновлён\n\nКарта: •••• %s\nНовый лимит: $%s/день", last4, limit.String()))
	}

	return nil
}

// GetUserSpendStats returns spend totals grouped by card category for the last 30 days.
func GetUserSpendStats(userID int) ([]map[string]interface{}, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	query := `
		SELECT COALESCE(c.category, 'arbitrage') as category,
		       COALESCE(SUM(t.amount), 0) as total_spent,
		       COUNT(t.id) as tx_count
		FROM transactions t
		LEFT JOIN cards c ON t.card_id = c.id
		WHERE t.user_id = $1
		  AND t.transaction_type IN ('CAPTURE', 'ISSUE')
		  AND t.executed_at >= NOW() - INTERVAL '30 days'
		GROUP BY COALESCE(c.category, 'arbitrage')
		ORDER BY total_spent DESC
	`
	rows, err := GlobalDB.Query(query, userID)
	if err != nil {
		log.Printf("DB Error GetUserSpendStats for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to fetch spend stats")
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var category string
		var totalSpent decimal.Decimal
		var txCount int
		if err := rows.Scan(&category, &totalSpent, &txCount); err != nil {
			log.Printf("DB Error scanning spend stats: %v", err)
			continue
		}
		stats = append(stats, map[string]interface{}{
			"category":    category,
			"total_spent": totalSpent.String(),
			"tx_count":    txCount,
		})
	}
	if stats == nil {
		stats = []map[string]interface{}{}
	}
	return stats, nil
}

// cryptoRand4 generates a cryptographically random 4-digit string (0000–9999).
func cryptoRand4() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "0000"
	}
	return fmt.Sprintf("%04d", n.Int64())
}

// IssueCards — Mock-провайдер для выпуска виртуальных карт (песочница).
// Генерирует карты с BIN 4455, случайным номером, CVV и Exp Date.
func IssueCards(userID int, req domain.MassIssueRequest) (interface{}, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	log.Printf("IssueCards: User %d requested %d cards", userID, req.Count)

	var results []domain.CardIssueResult
	successCount := 0
	failedCount := 0

	// Fee per card (calculated in handler and passed via context; also compute here for tx record)
	feePerCard := decimal.NewFromFloat(5.00)
	cat := req.Category
	if cat == "" {
		cat = "arbitrage"
	}
	switch cat {
	case "travel":
		feePerCard = decimal.NewFromFloat(3.00)
	case "services":
		feePerCard = decimal.NewFromFloat(2.00)
	}

	for i := 0; i < req.Count; i++ {
		// Генерируем случайные последние 4 цифры карты (криптографически)
		last4 := cryptoRand4()

		// Генерируем provider_card_id: BIN + 6 random digits + last4 = 16-digit PAN
		mid6 := fmt.Sprintf("%s%s", cryptoRand4()[:3], cryptoRand4()[:3])
		providerCardID := "4455" + mid6 + last4

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
		category := req.Category
		if category == "" {
			category = "arbitrage"
		}
		currency := strings.ToUpper(req.Currency)
		if currency != "EUR" {
			currency = "USD" // Default to USD unless explicitly EUR
		}

		// Если указан team_id, проверяем доступ пользователя к команде
		if req.TeamID != nil && *req.TeamID > 0 {
			hasAccess, _, err := CheckTeamAccess(*req.TeamID, userID)
			if err != nil || !hasAccess {
				log.Printf("Access denied: User %d does not have access to team %d", userID, *req.TeamID)
				failedCount++
				results = append(results, domain.CardIssueResult{
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
			INSERT INTO cards (user_id, provider_card_id, bin, last_4_digits, card_status, nickname, service_slug, daily_spend_limit, failed_auth_count, card_type, card_balance, team_id, category, currency)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			RETURNING id, created_at
		`,
			userID,
			providerCardID,
			"445500",
			last4,
			"ACTIVE",
			req.CardNickname,
			serviceSlug,
			req.DailyLimit,
			0,
			cardType,
			decimal.Zero,
			req.TeamID,
			category,
			currency,
		).Scan(&cardID, &createdAt)

		if err != nil {
			log.Printf("Failed to insert card for user %d: %v", userID, err)
			failedCount++
			results = append(results, domain.CardIssueResult{
				Success:   false,
				Status:    "FAILED",
				CardLast4: last4,
				Nickname:  req.CardNickname,
				Message:   fmt.Sprintf("Failed to issue card: %v", err),
			})
			continue
		}

		// Record transaction for card issuance (with actual fee)
		_, txErr := GlobalDB.Exec(
			`INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, executed_at)
			 VALUES ($1, $2, $3, $4, 'CARD_ISSUE', 'SUCCESS', $5, $6)`,
			userID, cardID, feePerCard, decimal.Zero,
			fmt.Sprintf("Card issued: %s •••• %s (%s) — fee $%s", cardType, last4, category, feePerCard.StringFixed(2)),
			time.Now(),
		)
		if txErr != nil {
			log.Printf("Warning: failed to record issue transaction for card %d: %v", cardID, txErr)
		}

		// Успешно создана карта
		successCount++
		results = append(results, domain.CardIssueResult{
			Success:   true,
			Status:    "ACTIVE",
			CardLast4: last4,
			Nickname:  req.CardNickname,
			Message:   "Card issued successfully",
			Card: &domain.Card{
				ID:              cardID,
				UserID:          userID,
				TeamID:          req.TeamID,
				BIN:             "445500",
				Last4Digits:     last4,
				CardStatus:      "ACTIVE",
				ServiceSlug:     serviceSlug,
				Category:        category,
				Currency:        currency,
				DailySpendLimit: req.DailyLimit,
				FailedAuthCount: 0,
				CardType:        cardType,
				CardBalance:     decimal.Zero,
				CreatedAt:       createdAt,
			},
		})
	}

	response := domain.MassIssueResponse{
		Successful: successCount,
		Failed:     failedCount,
		Results:    results,
	}

	log.Printf("✅ Issued %d cards successfully, %d failed for user %d", successCount, failedCount, userID)

	// Уведомление пользователю отправляется в handler (service.NotifyUser) — единое для TG+Email.
	// Здесь только уведомление админам.
	if successCount > 0 {
		// Уведомление админам
		go func() {
			email := ""
			if u, uErr := GetUserByID(userID); uErr == nil {
				email = u.Email
			}
			if email == "" {
				email = fmt.Sprintf("User #%d", userID)
			}
			adminMsg := fmt.Sprintf(
				"💳 <b>Выпуск карт</b>\n\n"+
					"👤 <b>Пользователь:</b> %s\n"+
					"📦 <b>Количество:</b> %d\n"+
					"🏷 <b>Категория:</b> %s\n"+
					"💰 <b>Комиссия:</b> $%s",
				email, successCount, cat, feePerCard.Mul(decimal.NewFromInt(int64(successCount))).StringFixed(2),
			)
			telegram.NotifyAdmins(adminMsg, "💳 Карты", "https://xplr.pro/admin/users")
		}()
	}

	// RevShare: 5% commission to referrer on card issuance
	if successCount > 0 {
		go func() {
			referrerID := GetReferrerID(userID)
			if referrerID <= 0 {
				return
			}
			issueAmount := decimal.NewFromInt(int64(successCount)).Mul(req.DailyLimit)
			if issueAmount.LessThanOrEqual(decimal.Zero) {
				issueAmount = decimal.NewFromInt(int64(successCount))
			}
			desc := fmt.Sprintf("%d cards issued (limit $%.2f each)", successCount, req.DailyLimit)
			if err := CreditRevShare(referrerID, userID, issueAmount, desc); err != nil {
				log.Printf("Warning: RevShare failed for user %d -> referrer %d: %v", userID, referrerID, err)
			}
		}()
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
func GetCardsNeedingReplenishment() ([]domain.Card, error) {
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

	var cards []domain.Card
	for rows.Next() {
		var card domain.Card
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
