package postgres

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type cardRepo struct {
	store *sqlx.DB
}

func NewCardRepository(db *sqlx.DB) ports.CardRepository {
	return &cardRepo{store: db}
}

func (r *cardRepo) Save(ctx context.Context, card *domain.Card) error {
	const query = `
		INSERT INTO cards (
			id, user_id, provider_card_id, bin, last_4_digits, card_status,
			nickname, daily_spend_limit, failed_auth_count, card_type,
			balance, expiry_date, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.store.ExecContext(ctx, query,
		card.ID, card.UserID, card.ProviderCardID, card.Bin, card.Last4Digits,
		card.CardStatus, card.Nickname, card.DailySpendLimit, card.FailedAuthCount,
		card.CardType, card.Balance, card.ExpiryDate, card.CreatedAt,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (r *cardRepo) GetByID(ctx context.Context, id domain.UUID) (*domain.Card, error) {
	const query = `
		SELECT 
			id, user_id, provider_card_id, bin, last_4_digits, card_status,
			nickname, daily_spend_limit, failed_auth_count, card_type,
			balance, expiry_date, created_at
		FROM cards 
		WHERE id = $1`

	var c domain.Card

	err := r.store.GetContext(ctx, &c, query, id)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &c, nil
}

func (r *cardRepo) Update(ctx context.Context, card *domain.Card) error {
	const query = `
		UPDATE cards 
		SET balance = $1, card_status = $2 
		WHERE id = $3`

	_, err := r.store.ExecContext(ctx, query, card.Balance, card.CardStatus, card.ID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
