package ports

import (
	"context"

	"github.com/djalben/xplr-core/internal/domain"
)

type CardRepository interface {
	Save(ctx context.Context, card *domain.Card) error
	GetByID(ctx context.Context, id domain.UUID) (*domain.Card, error)
	Update(ctx context.Context, card *domain.Card) error
	// ListByUserID и т.д. — добавим позже
}
