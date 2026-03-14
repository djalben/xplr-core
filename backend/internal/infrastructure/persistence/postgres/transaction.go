package postgres

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/internal/domain"
	"github.com/djalben/xplr-core/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type transactionRepo struct {
	store *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) ports.TransactionRepository {
	return &transactionRepo{store: db}
}

// Save — сохранение транзакции.
func (r *transactionRepo) Save(ctx context.Context, tx *domain.Transaction) error {
	const query = `
		INSERT INTO transactions (
			id, user_id, card_id, amount, fee, transaction_type, 
			status, details, provider_tx_id, executed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.store.ExecContext(ctx, query,
		tx.ID, tx.UserID, tx.CardID, tx.Amount, tx.Fee,
		tx.TransactionType, tx.Status, tx.Details,
		tx.ProviderTxID, tx.ExecutedAt,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// GetWalletTransactions — история по кошельку (card_id IS NULL).
func (r *transactionRepo) GetWalletTransactions(ctx context.Context, userID domain.UUID, from, to time.Time) ([]*domain.Transaction, error) {
	const query = `
		SELECT 
			id, user_id, card_id, amount, fee, transaction_type, 
			status, details, provider_tx_id, executed_at
		FROM transactions 
		WHERE user_id = $1 
		  AND card_id IS NULL 
		  AND executed_at BETWEEN $2 AND $3 
		ORDER BY executed_at DESC`

	var list []*domain.Transaction

	err := r.store.SelectContext(ctx, &list, query, userID, from, to)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return list, nil
}

// GetCardTransactions — история по одной карте.
func (r *transactionRepo) GetCardTransactions(ctx context.Context, cardID domain.UUID, from, to time.Time) ([]*domain.Transaction, error) {
	const query = `
		SELECT 
			id, user_id, card_id, amount, fee, transaction_type, 
			status, details, provider_tx_id, executed_at
		FROM transactions 
		WHERE card_id = $1 
		  AND executed_at BETWEEN $2 AND $3 
		ORDER BY executed_at DESC`

	var list []*domain.Transaction

	err := r.store.SelectContext(ctx, &list, query, cardID, from, to)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return list, nil
}
