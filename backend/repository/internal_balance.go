package repository

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/shopspring/decimal"
)

// GetInternalBalance — получить внутренний баланс (Кошелёк) пользователя.
// Если записи нет — создаёт с нулевым балансом (upsert).
func GetInternalBalance(userID int) (*models.InternalBalance, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		INSERT INTO internal_balances (user_id, master_balance, updated_at)
		VALUES ($1, 0, NOW())
		ON CONFLICT (user_id) DO NOTHING;

		SELECT id, user_id, master_balance, updated_at
		FROM internal_balances
		WHERE user_id = $1
	`

	// Выполняем upsert + select в два шага
	_, _ = GlobalDB.Exec(
		`INSERT INTO internal_balances (user_id, master_balance, updated_at)
		 VALUES ($1, 0, NOW()) ON CONFLICT (user_id) DO NOTHING`, userID,
	)

	var ib models.InternalBalance
	err := GlobalDB.QueryRow(
		`SELECT id, user_id, master_balance, COALESCE(auto_topup_enabled, FALSE), updated_at FROM internal_balances WHERE user_id = $1`,
		userID,
	).Scan(&ib.ID, &ib.UserID, &ib.MasterBalance, &ib.AutoTopupEnabled, &ib.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal balance: %w", err)
	}

	_ = query // suppress unused warning
	return &ib, nil
}

// TopUpInternalBalance — пополнить Кошелёк пользователя.
// Принимает сумму в рублях, конвертирует в USD по текущему курсу и зачисляет в master_balance (USD).
func TopUpInternalBalance(userID int, amountRub decimal.Decimal) (*models.InternalBalance, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	if amountRub.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Конвертируем RUB → USD по внутреннему курсу
	rate, err := GetFinalRate("RUB", "USD")
	if err != nil || rate.IsZero() {
		return nil, fmt.Errorf("exchange rate not available, cannot convert RUB to USD")
	}
	amountUsd := amountRub.Div(rate).Round(2)
	if amountUsd.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("converted amount too small")
	}

	tx, err := GlobalDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Зачисляем USD в Кошелёк (upsert)
	_, err = tx.Exec(
		`INSERT INTO internal_balances (user_id, master_balance, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (user_id) DO UPDATE SET master_balance = internal_balances.master_balance + $2, updated_at = NOW()`,
		userID, amountUsd,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to top up wallet: %w", err)
	}

	// Записываем транзакцию
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, amount, fee, transaction_type, status, details, executed_at)
		 VALUES ($1, $2, 0, 'WALLET_TOPUP', 'APPROVED', $3, $4)`,
		userID, amountUsd,
		fmt.Sprintf("Top-up wallet: %s ₽ → $%s (rate %s)", amountRub.StringFixed(0), amountUsd.StringFixed(2), rate.StringFixed(2)),
		time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Printf("✅ Wallet topped up: user=%d, %s ₽ → $%s", userID, amountRub.StringFixed(0), amountUsd.StringFixed(2))
	return GetInternalBalance(userID)
}

// DeductWalletBalance — списать из Кошелька (internal_balances.master_balance) для оплаты выпуска карт.
// Атомарно проверяет баланс, списывает и записывает транзакцию.
func DeductWalletBalance(userID int, amount decimal.Decimal, details string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("сумма списания должна быть положительной")
	}

	tx, err := GlobalDB.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer tx.Rollback()

	// Проверяем master_balance
	var balance decimal.Decimal
	err = tx.QueryRow(
		`SELECT COALESCE(master_balance, 0) FROM internal_balances WHERE user_id = $1 FOR UPDATE`,
		userID,
	).Scan(&balance)
	if err != nil {
		return fmt.Errorf("кошелёк не найден — пополните баланс")
	}
	if balance.LessThan(amount) {
		return fmt.Errorf("недостаточно средств (баланс: $%s, требуется: $%s)", balance.StringFixed(2), amount.StringFixed(2))
	}

	// Списываем
	_, err = tx.Exec(
		`UPDATE internal_balances SET master_balance = master_balance - $1, updated_at = NOW() WHERE user_id = $2`,
		amount, userID,
	)
	if err != nil {
		return fmt.Errorf("не удалось списать из кошелька: %v", err)
	}

	// Записываем транзакцию
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, amount, fee, transaction_type, status, details, executed_at)
		 VALUES ($1, $2, 0, 'CARD_ISSUE_FEE', 'APPROVED', $3, $4)`,
		userID, amount, details, time.Now(),
	)
	if err != nil {
		log.Printf("DB Error Insert deduction transaction: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка фиксации: %v", err)
	}

	log.Printf("User %d: deducted $%s from wallet for: %s", userID, amount.StringFixed(2), details)
	return nil
}

// TransferWalletToCard — перевести средства из Кошелька (USD) на карту.
// Если карта в EUR, конвертирует USD→EUR по внутреннему курсу.
// Атомарно: проверяет баланс, списывает из wallet, зачисляет на card_balance, записывает транзакцию.
func TransferWalletToCard(userID int, cardID int, amountInCardCurrency decimal.Decimal, cardCurrency string) (*models.InternalBalance, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	if amountInCardCurrency.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("сумма должна быть положительной")
	}

	// Определяем сколько USD нужно списать из кошелька
	var deductUSD decimal.Decimal
	currency := strings.ToUpper(strings.TrimSpace(cardCurrency))
	if currency == "" {
		currency = "USD"
	}

	if currency == "EUR" || currency == "€" {
		// Конвертируем EUR→USD: amountEUR * (rateRubPerEur / rateRubPerUsd)
		rateRubUsd, err := GetFinalRate("RUB", "USD")
		if err != nil || rateRubUsd.IsZero() {
			return nil, fmt.Errorf("курс USD недоступен")
		}
		rateRubEur, err := GetFinalRate("RUB", "EUR")
		if err != nil || rateRubEur.IsZero() {
			return nil, fmt.Errorf("курс EUR недоступен")
		}
		deductUSD = amountInCardCurrency.Mul(rateRubEur).Div(rateRubUsd).Round(2)
	} else {
		// USD → USD, 1:1
		deductUSD = amountInCardCurrency
	}

	tx, err := GlobalDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer tx.Rollback()

	// Проверяем master_balance и списываем
	var balance decimal.Decimal
	err = tx.QueryRow(
		`SELECT COALESCE(master_balance, 0) FROM internal_balances WHERE user_id = $1 FOR UPDATE`,
		userID,
	).Scan(&balance)
	if err != nil {
		return nil, fmt.Errorf("кошелёк не найден — пополните баланс")
	}
	if balance.LessThan(deductUSD) {
		return nil, fmt.Errorf("недостаточно средств (баланс: $%s, требуется: $%s)", balance.StringFixed(2), deductUSD.StringFixed(2))
	}

	_, err = tx.Exec(
		`UPDATE internal_balances SET master_balance = master_balance - $1, updated_at = NOW() WHERE user_id = $2`,
		deductUSD, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось списать из кошелька: %v", err)
	}

	// Проверяем принадлежность карты и зачисляем на card_balance
	var ownerID int
	err = tx.QueryRow(`SELECT user_id FROM cards WHERE id = $1 FOR UPDATE`, cardID).Scan(&ownerID)
	if err != nil {
		return nil, fmt.Errorf("карта не найдена")
	}
	if ownerID != userID {
		return nil, fmt.Errorf("нет доступа к этой карте")
	}

	_, err = tx.Exec(
		`UPDATE cards SET card_balance = COALESCE(card_balance, 0) + $1 WHERE id = $2`,
		amountInCardCurrency, cardID,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось зачислить на карту: %v", err)
	}

	// Записываем транзакцию
	sym := "$"
	if currency == "EUR" || currency == "€" {
		sym = "€"
	}
	details := fmt.Sprintf("Card top-up: %s%s → card #%d (deducted $%s from wallet)",
		sym, amountInCardCurrency.StringFixed(2), cardID, deductUSD.StringFixed(2))

	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, executed_at)
		 VALUES ($1, $2, $3, 0, 'CARD_TOPUP', 'APPROVED', $4, $5)`,
		userID, cardID, deductUSD, details, time.Now(),
	)
	if err != nil {
		log.Printf("DB Error Insert CARD_TOPUP transaction: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("ошибка фиксации: %v", err)
	}

	log.Printf("✅ User %d: transferred %s%s to card %d (deducted $%s from wallet)",
		userID, sym, amountInCardCurrency.StringFixed(2), cardID, deductUSD.StringFixed(2))

	return GetInternalBalance(userID)
}

// DeductInternalBalance — списать из Кошелька пользователя (вызывается Bridge при webhook).
// Атомарно уменьшает master_balance и увеличивает spent_from_wallet на карте.
func DeductInternalBalance(tx interface {
	Exec(string, ...interface{}) (interface {
		LastInsertId() (int64, error)
		RowsAffected() (int64, error)
	}, error)
	QueryRow(string, ...interface{}) interface{ Scan(...interface{}) error }
}, userID int, cardID int, amount decimal.Decimal) error {
	// Уменьшаем master_balance
	res := tx.QueryRow(
		`UPDATE internal_balances SET master_balance = master_balance - $1, updated_at = NOW()
		 WHERE user_id = $2 AND master_balance >= $1
		 RETURNING master_balance`,
		amount, userID,
	)
	var newBalance decimal.Decimal
	if err := res.Scan(&newBalance); err != nil {
		return fmt.Errorf("insufficient wallet balance or user not found")
	}

	// Увеличиваем spent_from_wallet на карте
	_, err := tx.Exec(
		`UPDATE cards SET spent_from_wallet = COALESCE(spent_from_wallet, 0) + $1 WHERE id = $2`,
		amount, cardID,
	)
	if err != nil {
		return fmt.Errorf("failed to update card spent_from_wallet: %w", err)
	}

	return nil
}

// ReclaimExpiredCardLimits — возвращает неиспользованный остаток лимита
// истёкших карт обратно в Кошелёк пользователя.
func ReclaimExpiredCardLimits() {
	if GlobalDB == nil {
		log.Println("[EXPIRY-RECLAIM] Database not initialized, skipping")
		return
	}

	log.Println("[EXPIRY-RECLAIM] Checking for expired cards...")

	// Находим карты с expiry_date в прошлом, у которых есть неиспользованный остаток лимита
	query := `
		SELECT c.id, c.user_id, c.spending_limit, COALESCE(c.spent_from_wallet, 0), c.last_4_digits
		FROM cards c
		WHERE c.expiry_date IS NOT NULL
		  AND c.expiry_date < NOW()
		  AND c.card_status != 'RECLAIMED'
		  AND COALESCE(c.spending_limit, 0) > COALESCE(c.spent_from_wallet, 0)
	`

	rows, err := GlobalDB.Query(query)
	if err != nil {
		log.Printf("[EXPIRY-RECLAIM] Error querying expired cards: %v", err)
		return
	}
	defer rows.Close()

	var reclaimed int
	for rows.Next() {
		var cardID, userID int
		var spendingLimit, spentFromWallet decimal.Decimal
		var last4 string

		if err := rows.Scan(&cardID, &userID, &spendingLimit, &spentFromWallet, &last4); err != nil {
			log.Printf("[EXPIRY-RECLAIM] Error scanning card: %v", err)
			continue
		}

		remainder := spendingLimit.Sub(spentFromWallet)
		if remainder.LessThanOrEqual(decimal.Zero) {
			continue
		}

		// Атомарно: вернуть остаток в Кошелёк + пометить карту как RECLAIMED
		tx, err := GlobalDB.Begin()
		if err != nil {
			log.Printf("[EXPIRY-RECLAIM] Error beginning tx: %v", err)
			continue
		}

		_, err = tx.Exec(
			`UPDATE internal_balances SET master_balance = master_balance + $1, updated_at = NOW() WHERE user_id = $2`,
			remainder, userID,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("[EXPIRY-RECLAIM] Error returning to wallet: %v", err)
			continue
		}

		_, err = tx.Exec(
			`UPDATE cards SET card_status = 'RECLAIMED', spending_limit = spent_from_wallet WHERE id = $1`,
			cardID,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("[EXPIRY-RECLAIM] Error marking card reclaimed: %v", err)
			continue
		}

		_, err = tx.Exec(
			`INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, executed_at)
			 VALUES ($1, $2, $3, 0, 'WALLET_RECLAIM', 'APPROVED', $4, $5)`,
			userID, cardID, remainder,
			fmt.Sprintf("Expired card ...%s: reclaimed %s back to wallet", last4, remainder.String()),
			time.Now(),
		)
		if err != nil {
			tx.Rollback()
			log.Printf("[EXPIRY-RECLAIM] Error recording tx: %v", err)
			continue
		}

		if err := tx.Commit(); err != nil {
			log.Printf("[EXPIRY-RECLAIM] Error committing: %v", err)
			continue
		}

		log.Printf("✅ [EXPIRY-RECLAIM] Card %d (user %d): reclaimed %s back to wallet", cardID, userID, remainder.String())
		reclaimed++
	}

	if reclaimed > 0 {
		log.Printf("[EXPIRY-RECLAIM] Reclaimed %d expired cards", reclaimed)
	} else {
		log.Println("[EXPIRY-RECLAIM] No expired cards to reclaim")
	}
}
