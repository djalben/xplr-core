package core

import (
	"fmt"
	"log"
	"time"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
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

	// 1. Проверить Кошелёк пользователя
	wallet, err := repository.GetInternalBalance(card.UserID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	// 2. Проверить баланс Кошелька
	if wallet.MasterBalance.LessThan(card.AutoReplenishAmount) {
		log.Printf("[AUTO-REPLENISH] Insufficient wallet balance for card %d. Wallet: $%s, Required: $%s",
			card.ID, wallet.MasterBalance.StringFixed(2), card.AutoReplenishAmount.StringFixed(2))

		go service.NotifyUser(card.UserID, "Автопополнение не удалось",
			fmt.Sprintf("⚠️ <b>Автопополнение не удалось</b>\n\n"+
				"Недостаточно средств для пополнения карты *%s.\n"+
				"Требуется: <b>$%s</b>, доступно: <b>$%s</b>\n\n"+
				"<a href=\"https://xplr.pro/wallet\">Пополнить кошелёк</a>",
				card.Last4Digits, card.AutoReplenishAmount.StringFixed(2), wallet.MasterBalance.StringFixed(2)))
		return fmt.Errorf("insufficient wallet balance")
	}

	// 3. Начать транзакцию БД для атомарности
	if repository.GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	tx, err := repository.GlobalDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 4. Списать из Кошелька (internal_balances.master_balance)
	_, err = tx.Exec(
		"UPDATE internal_balances SET master_balance = master_balance - $1, updated_at = NOW() WHERE user_id = $2 AND master_balance >= $1",
		card.AutoReplenishAmount, card.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to deduct from wallet: %w", err)
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

	// 8. Отправить уведомление пользователю (TG + Email)
	go service.NotifyUser(card.UserID, "Автопополнение карты",
		fmt.Sprintf("✅ <b>Автопополнение</b>\n\n"+
			"С вашего кошелька переведено <b>$%s</b> на карту *%s.\n"+
			"Новый баланс карты: <b>$%s</b>\n\n"+
			"<a href=\"https://xplr.pro/cards\">Открыть карты</a>",
			card.AutoReplenishAmount.StringFixed(2), card.Last4Digits,
			card.CardBalance.Add(card.AutoReplenishAmount).StringFixed(2)))

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
