package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type ExchangeRateRepository interface {
	ListAll(ctx context.Context) ([]*domain.ExchangeRate, error)
	Upsert(ctx context.Context, r *domain.ExchangeRate) error
}
