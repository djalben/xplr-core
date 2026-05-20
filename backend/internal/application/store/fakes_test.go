package store_test

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
)

type fakeWalletRepo struct {
	wallet      *domain.Wallet
	updateCalls int
	updateErr   error
}

func (f *fakeWalletRepo) GetByUserID(_ context.Context, _ domain.UUID) (*domain.Wallet, error) {
	return f.wallet, nil
}

func (f *fakeWalletRepo) EnsureWallet(_ context.Context, _ domain.UUID) error {
	return nil
}

func (f *fakeWalletRepo) Update(_ context.Context, w *domain.Wallet) error {
	if f.updateErr != nil {
		return f.updateErr
	}

	f.wallet = w
	f.updateCalls++

	return nil
}

type fakeStoreRepo struct {
	product   *domain.StoreProduct
	createErr error
}

func (f *fakeStoreRepo) ListCategories(context.Context) ([]*domain.StoreCategory, error) {
	return nil, nil
}

func (f *fakeStoreRepo) ListProducts(context.Context, ports.StoreProductFilter) ([]*domain.StoreProduct, error) {
	return nil, nil
}

func (f *fakeStoreRepo) GetProductByID(_ context.Context, id domain.UUID) (*domain.StoreProduct, error) {
	if f.product != nil && f.product.ID == id {
		return f.product, nil
	}

	return nil, domain.NewNotFound("product")
}

func (f *fakeStoreRepo) CreateOrder(_ context.Context, _ *domain.StoreOrder) error {
	return f.createErr
}

func (f *fakeStoreRepo) ListOrdersByUser(context.Context, domain.UUID, int) ([]*domain.StoreOrder, error) {
	return nil, nil
}

func (f *fakeStoreRepo) GetLatestCompletedOrderByProviderRef(context.Context, string) (*domain.StoreOrder, error) {
	return nil, domain.NewNotFound("order")
}

func (f *fakeStoreRepo) GetLatestCompletedOrderMetaByProviderRef(context.Context, string, *domain.UUID) (string, error) {
	return "", domain.NewNotFound("order")
}

func (f *fakeStoreRepo) SoftDeleteOrdersByProviderRef(context.Context, string) error {
	return nil
}

func (f *fakeStoreRepo) UpdateOrderMetaByProviderRef(context.Context, string, string) error {
	return nil
}

func (f *fakeStoreRepo) AdminListProducts(context.Context, ports.StoreAdminProductFilter) ([]*domain.StoreProduct, error) {
	return nil, nil
}

func (f *fakeStoreRepo) AdminUpdateProduct(context.Context, *domain.StoreProduct) error {
	return nil
}

func (f *fakeStoreRepo) AdminBulkAddMarkup(context.Context, domain.StoreProductType, domain.Numeric) (int64, error) {
	return 0, nil
}

func (f *fakeStoreRepo) AdminListVPNOrders(context.Context, int) ([]*domain.AdminVPNOrderRow, error) {
	return nil, nil
}

type fakeESIM struct {
	plan       *domain.ESIMPlan
	orderRes   *domain.ESIMOrderResult
	orderErr   error
	orderCalls int
}

func (f *fakeESIM) Name() string { return "fake" }

func (f *fakeESIM) GetDestinations(context.Context) ([]domain.ESIMDestination, error) {
	return nil, nil
}

func (f *fakeESIM) GetPlans(context.Context, string) ([]domain.ESIMPlan, error) {
	return nil, nil
}

func (f *fakeESIM) GetPlan(context.Context, string) (*domain.ESIMPlan, error) {
	return f.plan, nil
}

func (f *fakeESIM) OrderESIM(context.Context, string) (*domain.ESIMOrderResult, error) {
	f.orderCalls++

	return f.orderRes, f.orderErr
}

func (f *fakeESIM) CheckAvailability(context.Context, string) (bool, error) {
	return true, nil
}

type fakeVPN struct {
	key        string
	provider   string
	createErr  error
	createCalls int
}

func (f *fakeVPN) Name() string { return "fake" }

func (f *fakeVPN) CreateOrder(context.Context, string) (string, string, string, error) {
	f.createCalls++

	if f.createErr != nil {
		return "", "", "", f.createErr
	}

	return f.provider, f.key, "{}", nil
}

func (f *fakeVPN) GetClientTraffic(context.Context, string) (*domain.VPNClientTraffic, error) {
	return nil, nil
}

type fakeTxRepo struct {
	saved int
}

func (f *fakeTxRepo) Save(context.Context, *domain.Transaction) error {
	f.saved++

	return nil
}

func (f *fakeTxRepo) GetWalletTransactions(context.Context, domain.UUID, time.Time, time.Time) ([]*domain.Transaction, error) {
	return nil, nil
}

func (f *fakeTxRepo) GetCardTransactions(context.Context, domain.UUID, time.Time, time.Time) ([]*domain.Transaction, error) {
	return nil, nil
}

func (f *fakeTxRepo) GetByUserID(context.Context, domain.UUID, time.Time, time.Time, int) ([]*domain.Transaction, error) {
	return nil, nil
}

func (f *fakeTxRepo) SumCardSpendByCardID(context.Context, domain.UUID, time.Time, time.Time) (domain.Numeric, error) {
	return domain.NewNumeric(0), nil
}

func (f *fakeTxRepo) SumCardSpendByUserAndCardType(context.Context, domain.UUID, domain.CardType, time.Time, time.Time) (domain.Numeric, error) {
	return domain.NewNumeric(0), nil
}
