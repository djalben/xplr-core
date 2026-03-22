package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type ReferralRepository interface {
	CountByReferrer(ctx context.Context, referrerID domain.UUID) (int, error)
	TotalEarningsByReferrer(ctx context.Context, referrerID domain.UUID) (domain.Numeric, error)
}
