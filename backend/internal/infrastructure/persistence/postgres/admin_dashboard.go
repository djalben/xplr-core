package postgres

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type adminDashboardRepo struct {
	db *sqlx.DB
}

func NewAdminDashboardRepository(db *sqlx.DB) ports.AdminDashboardRepository {
	return &adminDashboardRepo{db: db}
}

func (r *adminDashboardRepo) GetStats(ctx context.Context) (*ports.AdminDashboardStats, error) {
	// NOTE: total_balance aggregated from wallets.balance, returned as text for exact precision.
	const q = `
SELECT
  (SELECT COUNT(*)::bigint FROM users) AS total_users,
  COALESCE((SELECT SUM(balance)::text FROM wallets), '0') AS total_balance,
  (SELECT COUNT(*)::bigint FROM cards WHERE card_status = 'ACTIVE') AS active_cards,
  (SELECT COUNT(*)::bigint FROM tickets WHERE status IN ('NEW','IN_PROGRESS')) AS open_tickets,
  (SELECT COUNT(*)::bigint FROM users WHERE created_at::date = CURRENT_DATE) AS today_signups,
  (SELECT COUNT(*)::bigint FROM cards) AS total_cards
`

	var out ports.AdminDashboardStats
	err := r.db.GetContext(ctx, &out, q)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &out, nil
}
