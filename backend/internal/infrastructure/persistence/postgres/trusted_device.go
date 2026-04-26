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

type trustedDeviceRepo struct {
	store *sqlx.DB
}

func NewTrustedDeviceRepository(db *sqlx.DB) ports.TrustedDeviceRepository {
	return &trustedDeviceRepo{store: db}
}

func (r *trustedDeviceRepo) Add(ctx context.Context, td *domain.TrustedDevice) error {
	if td == nil {
		return wrapper.Wrap(domain.NewInvalidInput("trusted device is required"))
	}
	if td.UserID == (domain.UUID{}) || td.TokenHash == "" || td.ExpiresAt.IsZero() {
		return wrapper.Wrap(domain.NewInvalidInput("user_id, token_hash and expires_at are required"))
	}

	const q = `
		INSERT INTO trusted_devices (
			id, user_id, token_hash, user_agent, ip, created_at, last_used_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	_, err := r.store.ExecContext(ctx, q,
		td.ID, td.UserID, td.TokenHash, td.UserAgent, td.IP, td.CreatedAt, td.LastUsedAt, td.ExpiresAt,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (r *trustedDeviceRepo) IsTrusted(ctx context.Context, userID domain.UUID, tokenHash string, now time.Time) (bool, error) {
	if userID == (domain.UUID{}) || tokenHash == "" {
		return false, wrapper.Wrap(domain.NewInvalidInput("user_id and token_hash are required"))
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	const q = `
		SELECT 1
		FROM trusted_devices
		WHERE user_id = $1
		  AND token_hash = $2
		  AND expires_at > $3
		LIMIT 1`

	var one int
	err := r.store.GetContext(ctx, &one, q, userID, tokenHash, now)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, wrapper.Wrap(err)
	}

	return true, nil
}

func (r *trustedDeviceRepo) TouchLastUsed(ctx context.Context, tokenHash string, now time.Time) error {
	if tokenHash == "" {
		return wrapper.Wrap(domain.NewInvalidInput("token_hash is required"))
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	const q = `UPDATE trusted_devices SET last_used_at = $1 WHERE token_hash = $2`
	_, err := r.store.ExecContext(ctx, q, now, tokenHash)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (r *trustedDeviceRepo) RevokeAll(ctx context.Context, userID domain.UUID) error {
	if userID == (domain.UUID{}) {
		return wrapper.Wrap(domain.NewInvalidInput("user_id is required"))
	}

	const q = `DELETE FROM trusted_devices WHERE user_id = $1`
	_, err := r.store.ExecContext(ctx, q, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
