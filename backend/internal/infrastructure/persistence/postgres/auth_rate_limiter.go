package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// Параметры анти-брутфорса (MVP).
// Login: окно 10 минут, максимум 10 попыток => блок 15 минут.
// MFA: окно 10 минут, максимум 6 попыток => блок 15 минут.
const (
	loginWindow   = 10 * time.Minute
	loginMaxFails = 10
	mfaWindow     = 10 * time.Minute
	mfaMaxFails   = 6
	blockDuration = 15 * time.Minute
	maxKeyLen     = 256
)

type authRateLimiter struct {
	db *sqlx.DB
}

func NewAuthRateLimiter(db *sqlx.DB) ports.AuthRateLimiter {
	return &authRateLimiter{db: db}
}

func (l *authRateLimiter) Allow(ctx context.Context, key string, now time.Time) (bool, time.Duration, error) {
	key = normalizeKey(key)
	if key == "" {
		return false, 0, wrapper.Wrap(domain.NewInvalidInput("rate limit key is required"))
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	const q = `SELECT blocked_until FROM auth_rate_limits WHERE key = $1`
	var blocked *time.Time
	err := l.db.GetContext(ctx, &blocked, q, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, 0, nil
		}

		return false, 0, wrapper.Wrap(err)
	}

	if blocked != nil && now.Before(*blocked) {
		return false, blocked.Sub(now), nil
	}

	return true, 0, nil
}

func (l *authRateLimiter) Fail(ctx context.Context, key string, now time.Time) (time.Duration, error) {
	key = normalizeKey(key)
	if key == "" {
		return 0, wrapper.Wrap(domain.NewInvalidInput("rate limit key is required"))
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	maxFails := maxFailsForKey(key)
	windowEnds := now.Add(loginWindow) // loginWindow == mfaWindow (оба 10m), отдельная функция не нужна

	// 1) если записи нет — создаём.
	const insert = `
		INSERT INTO auth_rate_limits (key, attempts, window_ends_at, blocked_until, updated_at)
		VALUES ($1, 1, $2, NULL, $3)
		ON CONFLICT (key) DO NOTHING`
	_, err := l.db.ExecContext(ctx, insert, key, windowEnds, now)
	if err != nil {
		return 0, wrapper.Wrap(err)
	}

	// 2) увеличиваем attempts; если окно истекло — сбрасываем на 1 и начинаем новое окно.
	const upd = `
		UPDATE auth_rate_limits
		SET
			attempts = CASE WHEN window_ends_at <= $2 THEN 1 ELSE attempts + 1 END,
			window_ends_at = CASE WHEN window_ends_at <= $2 THEN $3 ELSE window_ends_at END,
			updated_at = $2
		WHERE key = $1
		RETURNING attempts, window_ends_at, blocked_until`

	var attempts int
	var endsAt time.Time
	var blocked *time.Time
	err = l.db.QueryRowxContext(ctx, upd, key, now, windowEnds).Scan(&attempts, &endsAt, &blocked)
	if err != nil {
		return 0, wrapper.Wrap(err)
	}

	if blocked != nil && now.Before(*blocked) {
		return blocked.Sub(now), nil
	}

	if attempts >= maxFails {
		until := now.Add(blockDuration)
		const block = `UPDATE auth_rate_limits SET blocked_until = $2, updated_at = $3 WHERE key = $1`
		_, err = l.db.ExecContext(ctx, block, key, until, now)
		if err != nil {
			return 0, wrapper.Wrap(err)
		}

		return until.Sub(now), nil
	}

	return 0, nil
}

func (l *authRateLimiter) Success(ctx context.Context, key string, _ time.Time) error {
	key = normalizeKey(key)
	if key == "" {
		return wrapper.Wrap(domain.NewInvalidInput("rate limit key is required"))
	}

	const q = `DELETE FROM auth_rate_limits WHERE key = $1`
	_, err := l.db.ExecContext(ctx, q, key)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func normalizeKey(key string) string {
	if len(key) == 0 {
		return ""
	}
	if len(key) > maxKeyLen {
		return key[:maxKeyLen]
	}

	return key
}

func maxFailsForKey(key string) int {
	if len(key) >= 4 && key[:4] == "mfa:" {
		return mfaMaxFails
	}

	return loginMaxFails
}
