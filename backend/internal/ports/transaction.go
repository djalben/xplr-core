package ports

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type TransactionRepository interface {
	Save(ctx context.Context, tx *domain.Transaction) error

	// История по кошельку (card_id IS NULL)
	GetWalletTransactions(ctx context.Context, userID domain.UUID, from, to time.Time) ([]*domain.Transaction, error)

	// История по конкретной карте
	GetCardTransactions(ctx context.Context, cardID domain.UUID, from, to time.Time) ([]*domain.Transaction, error)

	// История по пользователю (все транзакции)
	GetByUserID(ctx context.Context, userID domain.UUID, from, to time.Time, limit int) ([]*domain.Transaction, error)

	// SumCardSpendByCardID — сумма CARD_SPEND по одной карте за полуинтервал [from, end).
	SumCardSpendByCardID(ctx context.Context, cardID domain.UUID, from, end time.Time) (domain.Numeric, error)

	// SumCardSpendByUserAndCardType — сумма CARD_SPEND по всем картам пользователя данного типа за [from, end).
	SumCardSpendByUserAndCardType(ctx context.Context, userID domain.UUID, cardType domain.CardType, from, end time.Time) (domain.Numeric, error)
}
