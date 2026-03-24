package ticket

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=contracts.go -destination=./mocks/contracts_mock.go -package=mocks

// TicketUseCase — сценарии HTTP-слоя /ticket (gomock).
type TicketUseCase interface {
	Create(ctx context.Context, userID domain.UUID, subject, message string, tgChatID *int64) (*domain.Ticket, error)
	Take(ctx context.Context, ticketID domain.UUID, adminID domain.UUID) error
	Close(ctx context.Context, ticketID domain.UUID, reply string) error
}
