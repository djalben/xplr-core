package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type NewsRepository interface {
	ListPublished(ctx context.Context, limit, offset int) ([]*domain.NewsArticle, int, error)
	ListAll(ctx context.Context, limit int) ([]*domain.NewsArticle, error)
	GetByID(ctx context.Context, id domain.UUID) (*domain.NewsArticle, error)
	Create(ctx context.Context, a *domain.NewsArticle) error
	Update(ctx context.Context, a *domain.NewsArticle) error
	SetStatus(ctx context.Context, id domain.UUID, status domain.NewsStatus) error
	Delete(ctx context.Context, id domain.UUID) error
}
