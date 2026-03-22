package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type CardRepository interface {
	Save(ctx context.Context, card *domain.Card) error
	GetByID(ctx context.Context, id domain.UUID) (*domain.Card, error)
	ListByUserID(ctx context.Context, userID domain.UUID) ([]*domain.Card, error)
	Update(ctx context.Context, card *domain.Card) error
}