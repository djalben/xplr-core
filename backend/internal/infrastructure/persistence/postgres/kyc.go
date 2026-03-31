package postgres

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type kycRepo struct {
	store *sqlx.DB
}

func NewKYCApplicationRepository(db *sqlx.DB) ports.KYCApplicationRepository {
	return &kycRepo{store: db}
}

func (r *kycRepo) Save(ctx context.Context, app *domain.KYCApplication) error {
	const query = `
		INSERT INTO kyc_applications (id, user_id, status, payload_json, admin_comment, reviewed_by, reviewed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.store.ExecContext(ctx, query,
		app.ID, app.UserID, app.Status, app.PayloadJSON, app.AdminComment,
		app.ReviewedBy, app.ReviewedAt, app.CreatedAt,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (r *kycRepo) GetByID(ctx context.Context, id domain.UUID) (*domain.KYCApplication, error) {
	const query = `
		SELECT id, user_id, status, payload_json, admin_comment, reviewed_by, reviewed_at, created_at
		FROM kyc_applications WHERE id = $1`

	var a domain.KYCApplication

	err := r.store.GetContext(ctx, &a, query, id)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &a, nil
}

func (r *kycRepo) ListByStatus(ctx context.Context, status domain.KYCApplicationStatus, limit int) ([]*domain.KYCApplication, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	const query = `
		SELECT id, user_id, status, payload_json, admin_comment, reviewed_by, reviewed_at, created_at
		FROM kyc_applications WHERE status = $1 ORDER BY created_at DESC LIMIT $2`

	var list []*domain.KYCApplication

	err := r.store.SelectContext(ctx, &list, query, status, limit)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return list, nil
}

func (r *kycRepo) HasPendingForUser(ctx context.Context, userID domain.UUID) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM kyc_applications WHERE user_id = $1 AND status = 'PENDING')`

	var ok bool

	err := r.store.GetContext(ctx, &ok, query, userID)
	if err != nil {
		return false, wrapper.Wrap(err)
	}

	return ok, nil
}

func (r *kycRepo) Update(ctx context.Context, app *domain.KYCApplication) error {
	const query = `
		UPDATE kyc_applications SET
			status = $1, payload_json = $2, admin_comment = $3, reviewed_by = $4, reviewed_at = $5
		WHERE id = $6`

	_, err := r.store.ExecContext(ctx, query,
		app.Status, app.PayloadJSON, app.AdminComment, app.ReviewedBy, app.ReviewedAt, app.ID,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
