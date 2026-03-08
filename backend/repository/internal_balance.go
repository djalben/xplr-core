package repository

import (
	"fmt"
	"log"
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
		`SELECT id, user_id, master_balance, updated_at FROM internal_balances WHERE user_id = $1`,
		userID,
	).Scan(&ib.ID, &ib.UserID, &ib.MasterBalance, &ib.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal balance: %w", err)
	}

	_ = query // suppress unused warning
	return &ib, nil
}

// TopUpInternalBalance — пополнить Кошелёк пользователя на указанную сумму.
// Списывает с balance_rub пользователя и зачисляет в internal_balances.
func TopUpInternalBalance(userID int, amount decimal.Decimal) (*models.InternalBalance, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amount must be positive")
	}

	tx, err := GlobalDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем достаточность balance_rub
	var balanceRub decimal.Decimal
	err = tx.QueryRow("SELECT COALESCE(balance_rub, 0) FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&balanceRub)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}
	if balanceRub.LessThan(amount) {
		return nil, fmt.Errorf("insufficient funds: available %s, required %s", balanceRub.String(), amount.String())
	}

	// Списываем с balance_rub
	_, err = tx.Exec("UPDATE users SET balance_rub = balance_rub - $1 WHERE id = $2", amount, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct user balance: %w", err)
	}

	// Зачисляем в Кошелёк (upsert)
	_, err = tx.Exec(
		`INSERT INTO internal_balances (user_id, master_balance, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (user_id) DO UPDATE SET master_balance = internal_balances.master_balance + $2, updated_at = NOW()`,
		userID, amount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to top up vault: %w", err)
	}

	// Записываем транзакцию
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, amount, fee, transaction_type, status, details, executed_at)
		 VALUES ($1, $2, 0, 'VAULT_TOPUP', 'APPROVED', $3, $4)`,
		userID, amount,
		fmt.Sprintf("Top-up vault: +%s", amount.String()),
		time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	log.Printf("✅ Vault topped up: user=%d, amount=%s", userID, amount.String())
	return GetInternalBalance(userID)
}

// DeductInternalBalance — списать из Кошелька пользователя (вызывается Bridge при webhook).
// Атомарно уменьшает master_balance и увеличивает spent_from_vault на карте.
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
		return fmt.Errorf("insufficient vault balance or user not found")
	}

	// Увеличиваем spent_from_vault на карте
	_, err := tx.Exec(
		`UPDATE cards SET spent_from_vault = COALESCE(spent_from_vault, 0) + $1 WHERE id = $2`,
		amount, cardID,
	)
	if err != nil {
		return fmt.Errorf("failed to update card spent_from_vault: %w", err)
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
		SELECT c.id, c.user_id, c.spending_limit, COALESCE(c.spent_from_vault, 0), c.last_4_digits
		FROM cards c
		WHERE c.expiry_date IS NOT NULL
		  AND c.expiry_date < NOW()
		  AND c.card_status != 'RECLAIMED'
		  AND COALESCE(c.spending_limit, 0) > COALESCE(c.spent_from_vault, 0)
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
		var spendingLimit, spentFromVault decimal.Decimal
		var last4 string

		if err := rows.Scan(&cardID, &userID, &spendingLimit, &spentFromVault, &last4); err != nil {
			log.Printf("[EXPIRY-RECLAIM] Error scanning card: %v", err)
			continue
		}

		remainder := spendingLimit.Sub(spentFromVault)
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
			log.Printf("[EXPIRY-RECLAIM] Error returning to vault: %v", err)
			continue
		}

		_, err = tx.Exec(
			`UPDATE cards SET card_status = 'RECLAIMED', spending_limit = spent_from_vault WHERE id = $1`,
			cardID,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("[EXPIRY-RECLAIM] Error marking card reclaimed: %v", err)
			continue
		}

		_, err = tx.Exec(
			`INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, executed_at)
			 VALUES ($1, $2, $3, 0, 'VAULT_RECLAIM', 'APPROVED', $4, $5)`,
			userID, cardID, remainder,
			fmt.Sprintf("Expired card ...%s: reclaimed %s back to vault", last4, remainder.String()),
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

		log.Printf("✅ [EXPIRY-RECLAIM] Card %d (user %d): reclaimed %s back to vault", cardID, userID, remainder.String())
		reclaimed++
	}

	if reclaimed > 0 {
		log.Printf("[EXPIRY-RECLAIM] Reclaimed %d expired cards", reclaimed)
	} else {
		log.Println("[EXPIRY-RECLAIM] No expired cards to reclaim")
	}
}
