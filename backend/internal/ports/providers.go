package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type ESIMProvider interface {
	Name() string
	GetDestinations(ctx context.Context) ([]domain.ESIMDestination, error)
	GetPlans(ctx context.Context, countryCode string) ([]domain.ESIMPlan, error)
	OrderESIM(ctx context.Context, planID string) (*domain.ESIMOrderResult, error)
	CheckAvailability(ctx context.Context, planID string) (bool, error)
}

type VPNProvider interface {
	Name() string
	CreateOrder(ctx context.Context, externalProductID string) (providerRef string, activationKey string, meta string, err error)
	GetClientTraffic(ctx context.Context, providerRef string) (*domain.VPNClientTraffic, error)
}
