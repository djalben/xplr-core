package ports

import (
	"context"
)

type AdminDashboardStats struct {
	TotalUsers   int64  `json:"totalUsers" db:"total_users"`
	TotalBalance string `json:"totalBalance" db:"total_balance"`
	ActiveCards  int64  `json:"activeCards" db:"active_cards"`
	OpenTickets  int64  `json:"openTickets" db:"open_tickets"`
	TodaySignups int64  `json:"todaySignups" db:"today_signups"`
	TotalCards   int64  `json:"totalCards" db:"total_cards"`
}

type AdminDashboardRepository interface {
	GetStats(ctx context.Context) (*AdminDashboardStats, error)
}
