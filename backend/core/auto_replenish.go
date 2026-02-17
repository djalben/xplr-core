package core

import (
	"fmt"
	"log"
	"time"

	"github.com/aalabin/xplr/models"
	"github.com/aalabin/xplr/notification"
	"github.com/aalabin/xplr/repository"
	"github.com/shopspring/decimal"
)

// ProcessAutoReplenishment - Проверяет и пополняет карты автоматически
func ProcessAutoReplenishment() {
	log.Println("[AUTO-REPLENISH] Starting auto-replenishment check...")

	// 1. Получить все карты, требующие пополнения
	cards, err := repository.GetCardsNeedingReplenishment()
	if err != nil {
		log.Printf("[AUTO-REPLENISH] Error fetching cards: %v", err)
		return
	}

	if len(cards) == 0 {
		log.Println("[AUTO-REPLENISH] No cards need replenishment")
		return
	}

	log.Printf("[AUTO-REPLENISH] Found %d cards needing replenishment", len(cards))

	// 2. Для каждой карты обработать пополнение
	for _, card := range cards {
		if err := processCardReplenishment(card); err != nil {
			log.Printf("[AUTO-REPLENISH] Error processing card %d: %v", card.ID, err)
			continue
		}
	}

	log.Println("[AUTO-REPLENISH] Auto-replenishment check completed")
}

// processCardReplenishment - Обработать пополнение одной карты
func processCardReplenishment(card models.Card) error {
	log.Printf("[AUTO-REPLENISH] Processing card %d (user %d, balance: %s, threshold: %s)",
		card.ID, card.UserID, card.CardBalance.String(), card.AutoReplenishThreshold.String())

	// 1. Получить пользователя
	user, err := repository.GetUserByID(card.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Проверить баланс пользователя
	if user.BalanceRub.LessThan(card.AutoReplenishAmount) {
		log.Printf("[AUTO-REPLENISH] Insufficient user balance for card %d. User balance_rub: %s, Required: %s",
			card.ID, user.BalanceRub.String(), card.AutoReplenishAmount.String())

		if user.TelegramChatID.Valid {
			notification.SendTelegramMessage(user.TelegramChatID.Int64,
				fmt.Sprintf("⚠️ *Auto-replenishment failed:* Недостаточно средств для пополнения карты `...%s`. Требуется: `%s`, доступно: `%s`",
					card.Last4Digits, card.AutoReplenishAmount.String(), user.BalanceRub.String()))
		}
		return fmt.Errorf("insufficient user balance")
	}

	// 3. Начать транзакцию БД для атомарности
	// Списать с баланса пользователя и пополнить карту
	if repository.GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	tx, err := repository.GlobalDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 4. Списать с баланса пользователя (XPLR: balance_rub)
	_, err = tx.Exec(
		"UPDATE users SET balance_rub = COALESCE(balance_rub, 0) - $1, balance = balance - $2 WHERE id = $3",
		card.AutoReplenishAmount, card.AutoReplenishAmount, card.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to deduct from user balance: %w", err)
	}

	// 5. Пополнить карту
	_, err = tx.Exec(
		"UPDATE cards SET card_balance = card_balance + $1 WHERE id = $2",
		card.AutoReplenishAmount, card.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to replenish card: %w", err)
	}

	// 6. Записать транзакцию FUND
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, executed_at)
		 VALUES ($1, $2, $3, $4, 'FUND', 'APPROVED', $5, $6)`,
		card.UserID,
		card.ID,
		card.AutoReplenishAmount,
		decimal.Zero, // Комиссия за автопополнение = 0
		fmt.Sprintf("Auto-replenishment: Card ...%s replenished with %s", card.Last4Digits, card.AutoReplenishAmount.String()),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	// 7. Коммит транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[AUTO-REPLENISH] ✅ Card %d replenished successfully with %s", card.ID, card.AutoReplenishAmount.String())

	// 8. Отправить уведомление пользователю
	if user.TelegramChatID.Valid {
		notification.SendTelegramMessage(user.TelegramChatID.Int64,
			fmt.Sprintf("✅ *Auto-replenishment:* Карта `...%s` пополнена на `%s`. Новый баланс карты: `%s`",
				card.Last4Digits, card.AutoReplenishAmount.String(), card.CardBalance.Add(card.AutoReplenishAmount).String()))
	}

	return nil
}

// StartAutoReplenishmentWorker - Запустить фоновый процесс автопополнения
func StartAutoReplenishmentWorker() {
	log.Println("[AUTO-REPLENISH] Starting auto-replenishment worker...")

	// Первый запуск через 1 минуту после старта сервера
	time.Sleep(1 * time.Minute)

	// Затем каждые 5 минут
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		// Первый запуск сразу
		ProcessAutoReplenishment()

		// Затем по расписанию
		for range ticker.C {
			ProcessAutoReplenishment()
		}
	}()

	log.Println("[AUTO-REPLENISH] Auto-replenishment worker started (checking every 5 minutes)")
}
