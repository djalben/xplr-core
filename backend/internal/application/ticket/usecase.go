package ticket

import (
	"context"

	"github.com/djalben/xplr-core/internal/domain"
	"github.com/djalben/xplr-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	ticketRepo ports.TicketRepository
}

func NewUseCase(tr ports.TicketRepository) *UseCase {
	return &UseCase{ticketRepo: tr}
}

func (uc *UseCase) Create(ctx context.Context, userID domain.UUID, subject, message string, tgChatID *int64) (*domain.Ticket, error) {
	t := domain.NewTicket(userID, subject, message, tgChatID)

	return t, wrapper.Wrap(uc.ticketRepo.Save(ctx, t))
}

func (uc *UseCase) Take(ctx context.Context, ticketID domain.UUID, adminID domain.UUID) error {
	t, err := uc.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	t.Take(adminID)

	return wrapper.Wrap(uc.ticketRepo.Update(ctx, t))
}

func (uc *UseCase) Close(ctx context.Context, ticketID domain.UUID, reply string) error {
	t, err := uc.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	t.Close(reply)

	return wrapper.Wrap(uc.ticketRepo.Update(ctx, t))
}
