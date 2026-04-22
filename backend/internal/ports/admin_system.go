package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type SystemSettingsRepository interface {
	ListAll(ctx context.Context) ([]*domain.SystemSetting, error)
	Upsert(ctx context.Context, s *domain.SystemSetting) error
}

type AdminLogsRepository interface {
	List(ctx context.Context, limit int) ([]*domain.AdminLog, error)
	Append(ctx context.Context, l *domain.AdminLog) error
}
