package postgres

import (
	"context"
	"database/sql"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type newsRepo struct {
	store *sqlx.DB
}

func NewNewsRepository(db *sqlx.DB) ports.NewsRepository {
	return &newsRepo{store: db}
}

func (r *newsRepo) ListPublished(ctx context.Context, limit, offset int) ([]*domain.NewsArticle, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var total int

	err := r.store.GetContext(ctx, &total, `SELECT COUNT(*) FROM news WHERE status = 'published'`)
	if err != nil {
		return nil, 0, wrapper.Wrap(err)
	}

	const q = `SELECT id, title, content, COALESCE(image_url, '') AS image_url, status, created_at, updated_at
		FROM news WHERE status = 'published' ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	var rows []*domain.NewsArticle

	err = r.store.SelectContext(ctx, &rows, q, limit, offset)
	if err != nil {
		return nil, 0, wrapper.Wrap(err)
	}

	if rows == nil {
		rows = []*domain.NewsArticle{}
	}

	return rows, total, nil
}

func (r *newsRepo) ListAll(ctx context.Context, limit int) ([]*domain.NewsArticle, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	const q = `SELECT id, title, content, COALESCE(image_url, '') AS image_url, status, created_at, updated_at
		FROM news ORDER BY created_at DESC LIMIT $1`

	var rows []*domain.NewsArticle

	err := r.store.SelectContext(ctx, &rows, q, limit)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	if rows == nil {
		rows = []*domain.NewsArticle{}
	}

	return rows, nil
}

func (r *newsRepo) GetByID(ctx context.Context, id domain.UUID) (*domain.NewsArticle, error) {
	const q = `SELECT id, title, content, COALESCE(image_url, '') AS image_url, status, created_at, updated_at FROM news WHERE id = $1`

	var a domain.NewsArticle

	err := r.store.GetContext(ctx, &a, q, id)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &a, nil
}

func (r *newsRepo) Create(ctx context.Context, a *domain.NewsArticle) error {
	const q = `INSERT INTO news (id, title, content, image_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.store.ExecContext(ctx, q, a.ID, a.Title, a.Content, a.ImageURL, a.Status, a.CreatedAt, a.UpdatedAt)

	return wrapper.Wrap(err)
}

func (r *newsRepo) Update(ctx context.Context, a *domain.NewsArticle) error {
	const q = `UPDATE news SET title = $1, content = $2, image_url = $3, status = $4, updated_at = $5 WHERE id = $6`

	_, err := r.store.ExecContext(ctx, q, a.Title, a.Content, a.ImageURL, a.Status, a.UpdatedAt, a.ID)

	return wrapper.Wrap(err)
}

func (r *newsRepo) SetStatus(ctx context.Context, id domain.UUID, status domain.NewsStatus) error {
	const q = `UPDATE news SET status = $1, updated_at = NOW() WHERE id = $2`

	_, err := r.store.ExecContext(ctx, q, status, id)

	return wrapper.Wrap(err)
}

func (r *newsRepo) Delete(ctx context.Context, id domain.UUID) error {
	const q = `DELETE FROM news WHERE id = $1`

	res, err := r.store.ExecContext(ctx, q, id)
	if err != nil {
		return wrapper.Wrap(err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return wrapper.Wrap(err)
	}
	if n == 0 {
		return wrapper.Wrap(sql.ErrNoRows)
	}

	return nil
}
