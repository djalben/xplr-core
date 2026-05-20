package store

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/google/uuid"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	storeRepo  ports.StoreRepository
	walletRepo ports.WalletRepository
	txRepo     ports.TransactionRepository
	esim       ports.ESIMProvider
	vpn        ports.VPNProvider
}

func NewUseCase(
	storeRepo ports.StoreRepository,
	walletRepo ports.WalletRepository,
	txRepo ports.TransactionRepository,
	esim ports.ESIMProvider,
	vpn ports.VPNProvider,
) *UseCase {
	return &UseCase{
		storeRepo:  storeRepo,
		walletRepo: walletRepo,
		txRepo:     txRepo,
		esim:       esim,
		vpn:        vpn,
	}
}

type CatalogResult struct {
	Categories []*domain.StoreCategory `json:"categories"`
	Products   []*domain.StoreProduct  `json:"products"`
}

func (uc *UseCase) Catalog(ctx context.Context, filter ports.StoreProductFilter) (*CatalogResult, error) {
	cats, err := uc.storeRepo.ListCategories(ctx)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	prods, err := uc.storeRepo.ListProducts(ctx, filter)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &CatalogResult{Categories: cats, Products: prods}, nil
}

type PurchaseResult struct {
	OrderID       domain.UUID `json:"orderId"`
	ProductName   string      `json:"productName"`
	PriceUSD      string      `json:"priceUsd"`
	ActivationKey string      `json:"activationKey"`
	QRData        string      `json:"qrData"`
	Status        string      `json:"status"`
	ProviderRef   string      `json:"providerRef"`
}

// ESIMPurchaseResult — ответ покупки eSIM с данными активации.
type ESIMPurchaseResult struct {
	StoreOrderID domain.UUID `json:"storeOrderId"`
	OrderID      string      `json:"orderId"`
	QRData       string      `json:"qrData"`
	LPA          string      `json:"lpa"`
	SMDP         string      `json:"smdp"`
	MatchingID   string      `json:"matchingId"`
	ICCID        string      `json:"iccid"`
	ProviderRef  string      `json:"providerRef"`
	Status       string      `json:"status"`
	PriceUSD     string      `json:"priceUsd"`
	ProductName  string      `json:"productName"`
}

func (uc *UseCase) Purchase(ctx context.Context, userID domain.UUID, productID domain.UUID) (*PurchaseResult, error) {
	p, err := uc.storeRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if !p.InStock {
		return nil, domain.NewInvalidInput("OUT_OF_STOCK")
	}

	pid := p.ID

	return uc.checkoutFromWallet(ctx, userID, p.PriceUSD, p.Name, &pid, func(ctx context.Context) (*orderFulfillment, error) {
		return uc.fulfillCatalogProduct(ctx, p)
	})
}

func (uc *UseCase) PurchaseESIM(ctx context.Context, userID domain.UUID, planID string) (*ESIMPurchaseResult, error) {
	if uc.esim == nil {
		return nil, domain.NewInvalidInput("PROVIDER_ERROR")
	}
	if planID == "" {
		return nil, domain.NewInvalidInput("planId is required")
	}

	plan, err := uc.esim.GetPlan(ctx, planID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if !plan.InStock {
		avail, availErr := uc.esim.CheckAvailability(ctx, planID)
		if availErr != nil || !avail {
			return nil, domain.NewInvalidInput("OUT_OF_STOCK")
		}
	}

	var esimRes *domain.ESIMOrderResult

	purchase, err := uc.checkoutFromWallet(ctx, userID, plan.PriceUSD, plan.Name, nil, func(ctx context.Context) (*orderFulfillment, error) {
		orderRes, orderErr := uc.esim.OrderESIM(ctx, planID)
		if orderErr != nil {
			return nil, domain.NewInvalidInput("PROVIDER_ERROR")
		}
		esimRes = orderRes

		meta := map[string]any{
			"order_id":    orderRes.OrderID,
			"lpa":         orderRes.LPA,
			"smdp":        orderRes.SMDP,
			"matching_id": orderRes.MatchingID,
			"iccid":       orderRes.ICCID,
			"plan_id":     planID,
			"provider":    plan.Provider,
		}
		status := domain.StoreOrderStatusCompleted
		if orderRes.PendingKYC {
			status = domain.StoreOrderStatusPending
			meta["kyc_url"] = orderRes.KYCURL
		}

		return &orderFulfillment{
			QRData:      orderRes.QRData,
			ProviderRef: orderRes.ProviderRef,
			Meta:        meta,
			Status:      status,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	providerOrderID := purchase.ProviderRef
	if esimRes != nil && esimRes.OrderID != "" {
		providerOrderID = esimRes.OrderID
	}

	out := &ESIMPurchaseResult{
		StoreOrderID: purchase.OrderID,
		OrderID:      providerOrderID,
		QRData:       purchase.QRData,
		ProviderRef:  purchase.ProviderRef,
		Status:       purchase.Status,
		PriceUSD:     purchase.PriceUSD,
		ProductName:  purchase.ProductName,
	}
	if esimRes != nil {
		out.LPA = esimRes.LPA
		out.SMDP = esimRes.SMDP
		out.MatchingID = esimRes.MatchingID
		out.ICCID = esimRes.ICCID
	}

	return out, nil
}

func (uc *UseCase) Orders(ctx context.Context, userID domain.UUID, limit int) ([]*domain.StoreOrder, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	return uc.storeRepo.ListOrdersByUser(ctx, userID, limit)
}

func (uc *UseCase) ESIMDestinations(ctx context.Context) ([]domain.ESIMDestination, error) {
	if uc.esim == nil {
		return []domain.ESIMDestination{}, nil
	}

	return uc.esim.GetDestinations(ctx)
}

func (uc *UseCase) ESIMPlans(ctx context.Context, countryCode string) ([]domain.ESIMPlan, error) {
	if uc.esim == nil {
		return []domain.ESIMPlan{}, nil
	}

	return uc.esim.GetPlans(ctx, countryCode)
}

func (uc *UseCase) fulfillCatalogProduct(ctx context.Context, p *domain.StoreProduct) (*orderFulfillment, error) {
	activationKey := ""
	qr := ""
	providerRef := "xplr-" + uuid.New().String()[:8]
	meta := map[string]any{}
	status := domain.StoreOrderStatusCompleted

	switch p.ProductType {
	case domain.StoreProductTypeVPN:
		if uc.vpn != nil {
			pRef, key, metaStr, err := uc.vpn.CreateOrder(ctx, p.ExternalID)
			if err != nil {
				return nil, domain.NewInvalidInput("PROVIDER_ERROR")
			}
			if key == "" {
				return nil, domain.NewInvalidInput("PROVIDER_ERROR")
			}

			providerRef = pRef
			activationKey = key
			meta["provider_meta"] = metaStr
		} else {
			activationKey = "vless://" + uuid.New().String()
			meta["traffic_bytes"] = int64(0)
			meta["expire_ms"] = time.Now().Add(time.Duration(p.ValidityDays) * 24 * time.Hour).UnixMilli()
			meta["duration_days"] = p.ValidityDays
		}
	case domain.StoreProductTypeESIM:
		if uc.esim == nil {
			return nil, domain.NewInvalidInput("PROVIDER_ERROR")
		}
		res, err := uc.esim.OrderESIM(ctx, p.ExternalID)
		if err != nil {
			return nil, domain.NewInvalidInput("PROVIDER_ERROR")
		}
		qr = res.QRData
		providerRef = res.ProviderRef
		meta["order_id"] = res.OrderID
		meta["lpa"] = res.LPA
		meta["smdp"] = res.SMDP
		meta["matching_id"] = res.MatchingID
		meta["iccid"] = res.ICCID
		if res.PendingKYC {
			status = domain.StoreOrderStatusPending
			meta["kyc_url"] = res.KYCURL
		}
	case domain.StoreProductTypeDigital:
		activationKey = "XPLR-DIG-" + uuid.New().String()[:12]
	}

	return &orderFulfillment{
		ActivationKey: activationKey,
		QRData:        qr,
		ProviderRef:   providerRef,
		Meta:          meta,
		Status:        status,
	}, nil
}
