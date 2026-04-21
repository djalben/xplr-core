package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type StoreRepository interface {
	ListCategories(ctx context.Context) ([]*domain.StoreCategory, error)
	ListProducts(ctx context.Context, filter StoreProductFilter) ([]*domain.StoreProduct, error)
	GetProductByID(ctx context.Context, id domain.UUID) (*domain.StoreProduct, error)
	CreateOrder(ctx context.Context, o *domain.StoreOrder) error
	ListOrdersByUser(ctx context.Context, userID domain.UUID, limit int) ([]*domain.StoreOrder, error)
	GetLatestCompletedOrderByProviderRef(ctx context.Context, providerRef string) (*domain.StoreOrder, error)
	GetLatestCompletedOrderMetaByProviderRef(ctx context.Context, providerRef string, userID *domain.UUID) (string, error)
	SoftDeleteOrdersByProviderRef(ctx context.Context, providerRef string) error
	UpdateOrderMetaByProviderRef(ctx context.Context, providerRef string, meta string) error
}

type StoreProductFilter struct {
	CategorySlug string
	Country      string
	Search       string
}
