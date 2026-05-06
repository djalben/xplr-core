package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type subscriptionRepo struct {
	store *sqlx.DB
}

func NewCardSubscriptionRepository(db *sqlx.DB) ports.CardSubscriptionRepository {
	return &subscriptionRepo{store: db}
}

func normalizeMerchantKey(merchantName string) string {
	return strings.ToLower(strings.TrimSpace(merchantName))
}

func (r *subscriptionRepo) UpsertOnCharge(
	ctx context.Context,
	userID domain.UUID,
	cardID domain.UUID,
	merchantName string,
	amount domain.Numeric,
	currency string,
	executedAt time.Time,
) (*domain.CardSubscription, error) {
	key := normalizeMerchantKey(merchantName)
	if key == "" {
		return nil, wrapper.Wrap(domain.NewInvalidInput("merchant_name is required"))
	}

	const query = `
		INSERT INTO card_subscriptions (
			id, user_id, card_id, merchant_name, merchant_key,
			last_amount, last_currency, charge_count,
			first_seen_at, last_seen_at,
			is_blocked, blocked_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, 1,
			$7, $7,
			FALSE, NULL
		)
		ON CONFLICT (card_id, merchant_key) DO UPDATE
		SET merchant_name = EXCLUDED.merchant_name,
		    last_amount = EXCLUDED.last_amount,
		    last_currency = EXCLUDED.last_currency,
		    charge_count = card_subscriptions.charge_count + 1,
		    last_seen_at = EXCLUDED.last_seen_at
		RETURNING
			id, user_id, card_id, merchant_name, merchant_key,
			last_amount, last_currency, charge_count,
			first_seen_at, last_seen_at,
			is_blocked, blocked_at`

	var out domain.CardSubscription
	err := r.store.GetContext(ctx, &out, query, userID, cardID, merchantName, key, amount, currency, executedAt)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &out, nil
}

func (r *subscriptionRepo) ListByUserID(ctx context.Context, userID domain.UUID) ([]*domain.CardSubscription, error) {
	const query = `
		SELECT
			id, user_id, card_id, merchant_name, merchant_key,
			last_amount, last_currency, charge_count,
			first_seen_at, last_seen_at,
			is_blocked, blocked_at
		FROM card_subscriptions
		WHERE user_id = $1
		ORDER BY last_seen_at DESC`

	var list []*domain.CardSubscription
	err := r.store.SelectContext(ctx, &list, query, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return list, nil
}

func (r *subscriptionRepo) SetBlocked(ctx context.Context, userID domain.UUID, subscriptionID domain.UUID, isBlocked bool) error {
	const query = `
		UPDATE card_subscriptions
		SET is_blocked = $1,
		    blocked_at = CASE WHEN $1 THEN NOW() ELSE NULL END
		WHERE id = $2 AND user_id = $3`

	_, err := r.store.ExecContext(ctx, query, isBlocked, subscriptionID, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (r *subscriptionRepo) SetBlockedByCardID(ctx context.Context, userID domain.UUID, cardID domain.UUID, isBlocked bool) error {
	const query = `
		UPDATE card_subscriptions
		SET is_blocked = $1,
		    blocked_at = CASE WHEN $1 THEN NOW() ELSE NULL END
		WHERE user_id = $2 AND card_id = $3`

	_, err := r.store.ExecContext(ctx, query, isBlocked, userID, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (r *subscriptionRepo) GetByCardAndMerchantKey(ctx context.Context, cardID domain.UUID, merchantKey string) (*domain.CardSubscription, error) {
	key := strings.ToLower(strings.TrimSpace(merchantKey))
	if key == "" {
		return nil, wrapper.Wrap(domain.NewInvalidInput("merchant_key is required"))
	}

	const query = `
		SELECT
			id, user_id, card_id, merchant_name, merchant_key,
			last_amount, last_currency, charge_count,
			first_seen_at, last_seen_at,
			is_blocked, blocked_at
		FROM card_subscriptions
		WHERE card_id = $1 AND merchant_key = $2`

	var out domain.CardSubscription
	err := r.store.GetContext(ctx, &out, query, cardID, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, wrapper.Wrap(domain.NewNotFound("subscription not found"))
		}

		return nil, wrapper.Wrap(err)
	}

	return &out, nil
}
