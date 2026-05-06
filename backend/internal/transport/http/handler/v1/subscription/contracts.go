package subscription

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=contracts.go -destination=./mocks/contracts_mock.go -package=mocks

type SubscriptionUseCase interface {
	List(ctx context.Context, userID domain.UUID) ([]*domain.CardSubscription, error)
	SetBlocked(ctx context.Context, userID domain.UUID, subscriptionID domain.UUID, isBlocked bool) error
	SetBlockedByCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, isBlocked bool) error
}

