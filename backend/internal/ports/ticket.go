package ports

import (
	"context"

	"github.com/djalben/xplr-core/internal/domain"
)

type TicketRepository interface {
	Save(ctx context.Context, ticket *domain.Ticket) error
	GetByID(ctx context.Context, id domain.UUID) (*domain.Ticket, error)
	Update(ctx context.Context, ticket *domain.Ticket) error
	// ListByUserID — добавим позже для админки
}
