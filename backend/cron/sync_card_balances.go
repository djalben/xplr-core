package cron

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/infrastructure/persistence/postgres"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// SyncCardBalances пересчитывает balance в таблице cards по данным из transactions.
// Правило:
// - TOPUP_CARD (COMPLETED) увеличивает баланс.
// - CARD_SPEND (COMPLETED) уменьшает баланс.
func SyncCardBalances(ctx context.Context, dsn string) (int64, error) {
	db, err := postgres.Connect(ctx, dsn)
	if err != nil {
		return 0, wrapper.Wrap(err)
	}
	defer db.Close()

	const query = `
		WITH
		topups AS (
			SELECT card_id, SUM(amount) AS sum_amount
			FROM transactions
			WHERE card_id IS NOT NULL
			  AND transaction_type = 'TOPUP_CARD'
			  AND status = 'COMPLETED'
			GROUP BY card_id
		),
		spends AS (
			SELECT card_id, SUM(amount) AS sum_amount
			FROM transactions
			WHERE card_id IS NOT NULL
			  AND transaction_type = 'CARD_SPEND'
			  AND status = 'COMPLETED'
			GROUP BY card_id
		)
		UPDATE cards c
		SET balance =
			COALESCE(t.sum_amount, 0) - COALESCE(s.sum_amount, 0)
		FROM topups t
		FULL OUTER JOIN spends s ON s.card_id = t.card_id
		WHERE c.id = COALESCE(t.card_id, s.card_id)`

	res, err := db.ExecContext(ctx, query)
	if err != nil {
		return 0, wrapper.Wrap(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, wrapper.Wrap(err)
	}

	return rows, nil
}
