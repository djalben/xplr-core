package postgres

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type exchangeRateRepo struct {
	db *sqlx.DB
}

func NewExchangeRateRepository(db *sqlx.DB) ports.ExchangeRateRepository {
	return &exchangeRateRepo{db: db}
}

func (r *exchangeRateRepo) ListAll(ctx context.Context) ([]*domain.ExchangeRate, error) {
	const q = `SELECT id, currency_from, currency_to, base_rate, markup_percent, final_rate, updated_at
FROM exchange_rates
ORDER BY currency_from, currency_to`

	var out []*domain.ExchangeRate
	err := r.db.SelectContext(ctx, &out, q)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if out == nil {
		out = []*domain.ExchangeRate{}
	}

	return out, nil
}

func (r *exchangeRateRepo) Upsert(ctx context.Context, er *domain.ExchangeRate) error {
	const q = `INSERT INTO exchange_rates (id, currency_from, currency_to, base_rate, markup_percent, final_rate, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (currency_from, currency_to)
DO UPDATE SET base_rate = EXCLUDED.base_rate, markup_percent = EXCLUDED.markup_percent, final_rate = EXCLUDED.final_rate, updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, q, er.ID, er.CurrencyFrom, er.CurrencyTo, er.BaseRate, er.MarkupPercent, er.FinalRate)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
