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

type authSessionsRepo struct {
	db *sqlx.DB
}

func NewAuthSessionsRepository(db *sqlx.DB) ports.AuthSessionsRepository {
	return &authSessionsRepo{db: db}
}

func (r *authSessionsRepo) Add(ctx context.Context, s *domain.AuthSession) error {
	if s == nil {
		return wrapper.Wrap(domain.NewInvalidInput("session is required"))
	}
	if s.UserID == (domain.UUID{}) {
		return wrapper.Wrap(domain.NewInvalidInput("user_id is required"))
	}

	if s.ID == (domain.UUID{}) {
		s.ID = domain.NewUUID()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}

	const q = `
		INSERT INTO auth_sessions (id, user_id, created_at, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, q, s.ID, s.UserID, s.CreatedAt, s.IP, s.UserAgent)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (r *authSessionsRepo) ListByUserID(ctx context.Context, userID domain.UUID, limit int) ([]*domain.AuthSession, error) {
	if userID == (domain.UUID{}) {
		return nil, wrapper.Wrap(domain.NewInvalidInput("user_id is required"))
	}
	if limit <= 0 {
		limit = 50
	}

	const q = `
		SELECT id, user_id, created_at, ip, user_agent
		FROM auth_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	var out []*domain.AuthSession
	err := r.db.SelectContext(ctx, &out, q, userID, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*domain.AuthSession{}, nil
		}

		return nil, wrapper.Wrap(err)
	}

	return out, nil
}

func (r *authSessionsRepo) DeleteByUserID(ctx context.Context, userID domain.UUID) error {
	if userID == (domain.UUID{}) {
		return wrapper.Wrap(domain.NewInvalidInput("user_id is required"))
	}

	const q = `DELETE FROM auth_sessions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, q, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (r *authSessionsRepo) DeleteOlderThan(ctx context.Context, userID domain.UUID, keepLast int) error {
	if userID == (domain.UUID{}) {
		return wrapper.Wrap(domain.NewInvalidInput("user_id is required"))
	}
	if keepLast <= 0 {
		return nil
	}

	const q = `
		DELETE FROM auth_sessions
		WHERE user_id = $1
		  AND id NOT IN (
			SELECT id
			FROM auth_sessions
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		  )`
	_, err := r.db.ExecContext(ctx, q, userID, keepLast)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
