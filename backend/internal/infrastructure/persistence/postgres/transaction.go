package postgres

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
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

// GetByUserID — все транзакции пользователя.
func (r *transactionRepo) GetByUserID(ctx context.Context, userID domain.UUID, from, to time.Time, limit int) ([]*domain.Transaction, error) {
	const query = `
		SELECT 
			id, user_id, card_id, amount, fee, transaction_type, 
			status, details, provider_tx_id, executed_at
		FROM transactions 
		WHERE user_id = $1 
		  AND executed_at BETWEEN $2 AND $3 
		ORDER BY executed_at DESC 
		LIMIT $4`

	if limit <= 0 {
		limit = 200
	}

	var list []*domain.Transaction

	err := r.store.SelectContext(ctx, &list, query, userID, from, to, limit)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return list, nil
}

// SumCardSpendByCardID — сумма успешных CARD_SPEND по карте за период.
func (r *transactionRepo) SumCardSpendByCardID(ctx context.Context, cardID domain.UUID, from, end time.Time) (domain.Numeric, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)::text
		FROM transactions
		WHERE card_id = $1
		  AND transaction_type = $2
		  AND status = 'COMPLETED'
		  AND executed_at >= $3 AND executed_at < $4`

	var s string

	err := r.store.GetContext(ctx, &s, query, cardID, domain.TransactionTypeCardSpend, from, end)
	if err != nil {
		return domain.NewNumeric(0), wrapper.Wrap(err)
	}

	sum, err := decimal.NewFromString(s)
	if err != nil {
		return domain.NewNumeric(0), wrapper.Wrap(err)
	}

	return sum, nil
}

// SumCardSpendByUserAndCardType — сумма CARD_SPEND по типу карты пользователя за период.
func (r *transactionRepo) SumCardSpendByUserAndCardType(
	ctx context.Context, userID domain.UUID, cardType domain.CardType, from, end time.Time,
) (domain.Numeric, error) {
	const query = `
		SELECT COALESCE(SUM(t.amount), 0)::text
		FROM transactions t
		INNER JOIN cards c ON c.id = t.card_id
		WHERE t.user_id = $1
		  AND c.card_type = $2
		  AND t.transaction_type = $3
		  AND t.status = 'COMPLETED'
		  AND t.executed_at >= $4 AND t.executed_at < $5`

	var s string

	err := r.store.GetContext(ctx, &s, query, userID, string(cardType), domain.TransactionTypeCardSpend, from, end)
	if err != nil {
		return domain.NewNumeric(0), wrapper.Wrap(err)
	}

	sum, err := decimal.NewFromString(s)
	if err != nil {
		return domain.NewNumeric(0), wrapper.Wrap(err)
	}

	return sum, nil
}
