package postgres

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type referralRepo struct {
	store *sqlx.DB
}

func NewReferralRepository(db *sqlx.DB) ports.ReferralRepository {
	return &referralRepo{store: db}
}

func (r *referralRepo) CountByReferrer(ctx context.Context, referrerID domain.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM referrals WHERE referrer_id = $1`

	var n int

	err := r.store.GetContext(ctx, &n, query, referrerID)
	if err != nil {
		return 0, wrapper.Wrap(err)
	}

	return n, nil
}

func (r *referralRepo) TotalEarningsByReferrer(ctx context.Context, referrerID domain.UUID) (domain.Numeric, error) {
	const query = `SELECT COALESCE(SUM(commission_earned), 0) FROM referrals WHERE referrer_id = $1`

	var sum float64

	err := r.store.GetContext(ctx, &sum, query, referrerID)
	if err != nil {
		return domain.NewNumeric(0), wrapper.Wrap(err)
	}

	return domain.NewNumeric(sum), nil
}
