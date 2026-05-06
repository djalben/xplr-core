package ports

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type CardSubscriptionRepository interface {
	UpsertOnCharge(
		ctx context.Context,
		userID domain.UUID,
		cardID domain.UUID,
		merchantName string,
		amount domain.Numeric,
		currency string,
		executedAt time.Time,
	) (*domain.CardSubscription, error)

	ListByUserID(ctx context.Context, userID domain.UUID) ([]*domain.CardSubscription, error)

	SetBlocked(ctx context.Context, userID domain.UUID, subscriptionID domain.UUID, isBlocked bool) error
	SetBlockedByCardID(ctx context.Context, userID domain.UUID, cardID domain.UUID, isBlocked bool) error

	GetByCardAndMerchantKey(ctx context.Context, cardID domain.UUID, merchantKey string) (*domain.CardSubscription, error)
}
