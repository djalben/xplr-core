package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

//nolint:interfacebloat // Repository aggregates user+admin store operations; split is a larger refactor.
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

	// Admin-only
	AdminListProducts(ctx context.Context, filter StoreAdminProductFilter) ([]*domain.StoreProduct, error)
	AdminUpdateProduct(ctx context.Context, p *domain.StoreProduct) error
	AdminBulkAddMarkup(ctx context.Context, productType domain.StoreProductType, delta domain.Numeric) (affected int64, err error)

	AdminListVPNOrders(ctx context.Context, limit int) ([]*domain.AdminVPNOrderRow, error)
}

type StoreProductFilter struct {
	CategorySlug string
	Country      string
	Search       string
}

type StoreAdminProductFilter struct {
	ProductType *domain.StoreProductType
}
