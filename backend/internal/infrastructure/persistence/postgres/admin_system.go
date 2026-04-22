package postgres

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type systemSettingsRepo struct {
	db *sqlx.DB
}

func NewSystemSettingsRepository(db *sqlx.DB) ports.SystemSettingsRepository {
	return &systemSettingsRepo{db: db}
}

func (r *systemSettingsRepo) ListAll(ctx context.Context) ([]*domain.SystemSetting, error) {
	const q = `SELECT setting_key, setting_value, description, updated_at FROM system_settings ORDER BY setting_key`

	var out []*domain.SystemSetting

	err := r.db.SelectContext(ctx, &out, q)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if out == nil {
		out = []*domain.SystemSetting{}
	}

	return out, nil
}

func (r *systemSettingsRepo) Upsert(ctx context.Context, s *domain.SystemSetting) error {
	const q = `INSERT INTO system_settings (setting_key, setting_value, description, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (setting_key) DO UPDATE SET setting_value = EXCLUDED.setting_value, description = EXCLUDED.description, updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, q, s.Key, s.Value, s.Description)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

type adminLogsRepo struct {
	db *sqlx.DB
}

func NewAdminLogsRepository(db *sqlx.DB) ports.AdminLogsRepository {
	return &adminLogsRepo{db: db}
}

func (r *adminLogsRepo) List(ctx context.Context, limit int) ([]*domain.AdminLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	const q = `SELECT id, admin_id, action, created_at FROM admin_logs ORDER BY created_at DESC LIMIT $1`

	var out []*domain.AdminLog

	err := r.db.SelectContext(ctx, &out, q, limit)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if out == nil {
		out = []*domain.AdminLog{}
	}

	return out, nil
}

func (r *adminLogsRepo) Append(ctx context.Context, l *domain.AdminLog) error {
	const q = `INSERT INTO admin_logs (id, admin_id, action, created_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, q, l.ID, l.AdminID, l.Action, l.CreatedAt)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
