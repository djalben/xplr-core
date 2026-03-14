package ports

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/internal/domain"
)

type TransactionRepository interface {
	Save(ctx context.Context, tx *domain.Transaction) error

	// История по кошельку (card_id IS NULL)
	GetWalletTransactions(ctx context.Context, userID domain.UUID, from, to time.Time) ([]*domain.Transaction, error)

	// История по конкретной карте
	GetCardTransactions(ctx context.Context, cardID domain.UUID, from, to time.Time) ([]*domain.Transaction, error)
}
