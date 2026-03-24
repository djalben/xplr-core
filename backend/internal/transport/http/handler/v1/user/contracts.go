package user

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=contracts.go -destination=./mocks/contracts_mock.go -package=mocks

// UserProfile — профиль и рефералы (HTTP user BFF).
type UserProfile interface {
	GetMe(ctx context.Context, userID domain.UUID) (map[string]any, error)
	GetReferralInfo(ctx context.Context, userID domain.UUID) (map[string]any, error)
}

// UserWallet — кошелёк в user-хендлере.
type UserWallet interface {
	GetBalance(ctx context.Context, userID domain.UUID) (domain.Numeric, error)
	TopUpWallet(ctx context.Context, userID domain.UUID, amount domain.Numeric) error
	ToggleAutoTopUp(ctx context.Context, userID domain.UUID, enabled bool) error
}

// UserGrades — грейд пользователя.
type UserGrades interface {
	GetByUserID(ctx context.Context, userID domain.UUID) (*domain.UserGrade, error)
}

// UserCards — карты в user BFF.
type UserCards interface {
	BuyCard(ctx context.Context, userID domain.UUID, cardType domain.CardType, nickname string) (*domain.Card, error)
	ListByUserID(ctx context.Context, userID domain.UUID) ([]*domain.Card, error)
	GetByID(ctx context.Context, cardID domain.UUID) (*domain.Card, error)
	TopUpCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, amount domain.Numeric) error
	UpdateStatus(ctx context.Context, userID domain.UUID, cardID domain.UUID, status string) error
	SetSpendingLimit(ctx context.Context, userID domain.UUID, cardID domain.UUID, limit domain.Numeric) error
}

// UserTransactions — единая лента транзакций.
type UserTransactions interface {
	GetUnifiedTransactions(ctx context.Context, userID domain.UUID, from, to time.Time, limit int) ([]*domain.Transaction, error)
}

// UserTickets — тикеты (support).
type UserTickets interface {
	Create(ctx context.Context, userID domain.UUID, subject, message string, tgChatID *int64) (*domain.Ticket, error)
}
