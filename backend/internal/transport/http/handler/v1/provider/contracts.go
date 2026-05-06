package provider

import (
	"context"

	subscriptionUC "github.com/djalben/xplr-core/backend/internal/application/subscription"
)

type SubscriptionAuthorizationUseCase interface {
	HandleAuthorization(ctx context.Context, event subscriptionUC.AuthorizationEvent) (subscriptionUC.AuthorizationResult, error)
}

