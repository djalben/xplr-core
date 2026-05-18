package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/application/store"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/google/uuid"
)

func testUID() domain.UUID {
	return uuid.MustParse("11111111-1111-1111-1111-111111111111")
}

func testWallet(balance float64) *domain.Wallet {
	w := domain.NewWallet(testUID())
	_ = w.TopUp(domain.NewNumeric(balance))

	return w
}

func TestPurchaseESIM_InsufficientFunds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()

	plan := &domain.ESIMPlan{
		PlanID:   "demo-tr-1gb",
		Name:     "Demo 1 GB",
		PriceUSD: domain.NewNumeric(10),
		InStock:  true,
	}

	uc := store.NewUseCase(
		&fakeStoreRepo{},
		&fakeWalletRepo{wallet: testWallet(5)},
		nil,
		&fakeESIM{plan: plan},
		nil,
	)

	_, err := uc.PurchaseESIM(ctx, uid, plan.PlanID)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domain.ErrInsufficientFunds) {
		t.Fatalf("want ErrInsufficientFunds, got %v", err)
	}
}

func TestPurchaseESIM_OK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()

	plan := &domain.ESIMPlan{
		PlanID:   "demo-tr-1gb",
		Provider: "demo",
		Name:     "Demo 1 GB",
		PriceUSD: domain.NewNumeric(3.5),
		InStock:  true,
	}
	orderRes := &domain.ESIMOrderResult{
		OrderID:     "MM-123",
		QRData:      "LPA:1$host$code",
		LPA:         "LPA:1$host$code",
		SMDP:        "host",
		MatchingID:  "code",
		ICCID:       "8901",
		ProviderRef: "MM-123",
	}

	uc := store.NewUseCase(
		&fakeStoreRepo{},
		&fakeWalletRepo{wallet: testWallet(20)},
		&fakeTxRepo{},
		&fakeESIM{plan: plan, orderRes: orderRes},
		nil,
	)

	out, err := uc.PurchaseESIM(ctx, uid, plan.PlanID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.QRData == "" {
		t.Fatal("expected qr data")
	}
	if out.StoreOrderID == uuid.Nil {
		t.Fatal("expected store order id")
	}
}

func TestPurchaseESIM_RefundOnCreateOrderFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()

	plan := &domain.ESIMPlan{
		PlanID:   "demo-tr-1gb",
		Name:     "Demo 1 GB",
		PriceUSD: domain.NewNumeric(3.5),
		InStock:  true,
	}
	orderRes := &domain.ESIMOrderResult{
		OrderID:     "MM-123",
		QRData:      "LPA:1$host$code",
		LPA:         "LPA:1$host$code",
		ProviderRef: "MM-123",
	}

	walletRepo := &fakeWalletRepo{wallet: testWallet(20)}
	uc := store.NewUseCase(
		&fakeStoreRepo{createErr: errors.New("db down")},
		walletRepo,
		nil,
		&fakeESIM{plan: plan, orderRes: orderRes},
		nil,
	)

	_, err := uc.PurchaseESIM(ctx, uid, plan.PlanID)
	if err == nil {
		t.Fatal("expected error")
	}
	if !walletRepo.wallet.Balance.Equal(domain.NewNumeric(20)) {
		t.Fatalf("wallet not refunded, balance=%s", walletRepo.wallet.Balance.String())
	}
}
