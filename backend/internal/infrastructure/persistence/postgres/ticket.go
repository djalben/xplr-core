package postgres

import (
	"context"

	"github.com/djalben/xplr-core/internal/domain"
	"github.com/djalben/xplr-core/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type ticketRepo struct {
	store *sqlx.DB
}

func NewTicketRepository(db *sqlx.DB) ports.TicketRepository {
	return &ticketRepo{store: db}
}

// Save — создание тикета.
func (r *ticketRepo) Save(ctx context.Context, ticket *domain.Ticket) error {
	const query = `
		INSERT INTO tickets (id, user_id, admin_id, tg_chat_id, subject, status, 
		                     user_message, admin_reply, created_at, closed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.store.ExecContext(ctx, query,
		ticket.ID, ticket.UserID, ticket.AdminID, ticket.TGChatID,
		ticket.Subject, ticket.Status, ticket.UserMessage,
		ticket.AdminReply, ticket.CreatedAt, ticket.ClosedAt,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// GetByID — получение тикета.
func (r *ticketRepo) GetByID(ctx context.Context, id domain.UUID) (*domain.Ticket, error) {
	const query = `
		SELECT id, user_id, admin_id, tg_chat_id, subject, status, 
		       user_message, admin_reply, created_at, closed_at
		FROM tickets 
		WHERE id = $1`

	var t domain.Ticket

	err := r.store.GetContext(ctx, &t, query, id)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &t, nil
}

// Update — обновление тикета.
func (r *ticketRepo) Update(ctx context.Context, ticket *domain.Ticket) error {
	const query = `
		UPDATE tickets 
		SET admin_id = $1, status = $2, admin_reply = $3, closed_at = $4
		WHERE id = $5`

	_, err := r.store.ExecContext(ctx, query,
		ticket.AdminID, ticket.Status, ticket.AdminReply,
		ticket.ClosedAt, ticket.ID,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
