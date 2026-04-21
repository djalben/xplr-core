package postgres

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type commissionRepo struct {
	store *sqlx.DB
}

func NewCommissionConfigRepository(db *sqlx.DB) ports.CommissionConfigRepository {
	return &commissionRepo{store: db}
}

// GetByKey — получение комиссии по ключу.
func (r *commissionRepo) GetByKey(ctx context.Context, key string) (*domain.CommissionConfig, error) {
	const query = `SELECT id, key, value, description, updated_at FROM commission_config WHERE key = $1`

	var cfg domain.CommissionConfig

	err := r.store.GetContext(ctx, &cfg, query, key)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &cfg, nil
}

// Update — обновление комиссии.
func (r *commissionRepo) Update(ctx context.Context, cfg *domain.CommissionConfig) error {
	const query = `UPDATE commission_config SET value = $1, description = $2, updated_at = NOW() WHERE key = $3`

	_, err := r.store.ExecContext(ctx, query, cfg.Value, cfg.Description, cfg.Key)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// ListAll — все комиссии (для админки).
func (r *commissionRepo) ListAll(ctx context.Context) ([]*domain.CommissionConfig, error) {
	const query = `SELECT id, key, value, description, updated_at FROM commission_config ORDER BY key`

	var list []*domain.CommissionConfig

	err := r.store.SelectContext(ctx, &list, query)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return list, nil
}
