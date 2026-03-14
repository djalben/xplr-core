package domain

import "time"

type TicketStatus string

const (
	TicketNew        TicketStatus = "NEW"
	TicketInProgress TicketStatus = "IN_PROGRESS"
	TicketDone       TicketStatus = "DONE"
)

type Ticket struct {
	ID          UUID         `json:"id" db:"id"`
	UserID      UUID         `json:"userId" db:"user_id"`
	AdminID     *UUID        `json:"adminId,omitempty" db:"admin_id"`
	TGChatID    *int64       `json:"tgChatId,omitempty" db:"tg_chat_id"`
	Subject     string       `json:"subject" db:"subject"`
	Status      TicketStatus `json:"status" db:"status"`
	UserMessage string       `json:"userMessage" db:"user_message"`
	AdminReply  string       `json:"adminReply,omitempty" db:"admin_reply"`
	CreatedAt   time.Time    `json:"createdAt" db:"created_at"`
	ClosedAt    *time.Time   `json:"closedAt,omitempty" db:"closed_at"`
}

func NewTicket(userID UUID, subject, message string, tgChatID *int64) *Ticket {
	return &Ticket{
		ID:          NewUUID(),
		UserID:      userID,
		Subject:     subject,
		Status:      TicketNew,
		UserMessage: message,
		TGChatID:    tgChatID,
		CreatedAt:   time.Now().UTC(),
	}
}

func (t *Ticket) Take(adminID UUID) {
	t.AdminID = &adminID
	t.Status = TicketInProgress
}

func (t *Ticket) Close(reply string) {
	t.AdminReply = reply
	t.Status = TicketDone
	now := time.Now().UTC()
	t.ClosedAt = &now
}
