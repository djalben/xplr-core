package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/google/uuid"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	storeRepo  ports.StoreRepository
	walletRepo ports.WalletRepository
	esim       ports.ESIMProvider
	vpn        ports.VPNProvider
}

func NewUseCase(storeRepo ports.StoreRepository, walletRepo ports.WalletRepository, esim ports.ESIMProvider, vpn ports.VPNProvider) *UseCase {
	return &UseCase{storeRepo: storeRepo, walletRepo: walletRepo, esim: esim, vpn: vpn}
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

var ErrNoActiveCard = domain.NewInvalidInput("NO_ACTIVE_CARD")

func (uc *UseCase) Purchase(ctx context.Context, userID domain.UUID, productID domain.UUID) (*PurchaseResult, error) {
	p, err := uc.storeRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if !p.InStock {
		return nil, wrapper.Wrap(domain.NewInvalidInput("OUT_OF_STOCK"))
	}

	// MVP: оплата из кошелька (в main это «через карту»). Для sandbox сейчас делаем так, чтобы цепочка работала.
	w, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if w.Balance.LessThan(p.PriceUSD) {
		return nil, domain.NewInsufficientFunds()
	}
	err = w.Withdraw(p.PriceUSD)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	err = uc.walletRepo.Update(ctx, w)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	activationKey := ""
	qr := ""
	providerRef := "xplr-" + uuid.New().String()[:8]

	meta := map[string]any{}
	if p.ProductType == domain.StoreProductTypeVPN {
		if uc.vpn != nil {
			pRef, key, metaStr, err := uc.vpn.CreateOrder(ctx, p.ExternalID)
			if err == nil {
				providerRef = pRef
				activationKey = key
				meta["provider_meta"] = metaStr
			}
		}

		// Fallback for dev environments without XPanel configured.
		if activationKey == "" {
			activationKey = "vless://" + uuid.New().String()
			meta["traffic_bytes"] = int64(0)
			meta["expire_ms"] = time.Now().Add(time.Duration(p.ValidityDays) * 24 * time.Hour).UnixMilli()
			meta["duration_days"] = p.ValidityDays
		}
	}
	if p.ProductType == domain.StoreProductTypeESIM {
		if uc.esim == nil {
			return nil, wrapper.Wrap(domain.NewInvalidInput("PROVIDER_ERROR"))
		}
		// For catalog-driven purchase, external_id stores plan ID.
		res, err := uc.esim.OrderESIM(ctx, p.ExternalID)
		if err != nil {
			return nil, wrapper.Wrap(domain.NewInvalidInput("PROVIDER_ERROR"))
		}
		qr = res.QRData
		providerRef = res.ProviderRef
		meta["order_id"] = res.OrderID
		meta["lpa"] = res.LPA
		meta["smdp"] = res.SMDP
		meta["matching_id"] = res.MatchingID
		meta["iccid"] = res.ICCID
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	oid := uuid.New()
	o := &domain.StoreOrder{
		ID:            oid,
		UserID:        userID,
		ProductID:     &p.ID,
		ProductName:   p.Name,
		PriceUSD:      p.PriceUSD,
		Status:        domain.StoreOrderStatusCompleted,
		ActivationKey: activationKey,
		QRData:        qr,
		ProviderRef:   providerRef,
		Meta:          string(metaBytes),
		CreatedAt:     time.Now().UTC(),
	}
	err = uc.storeRepo.CreateOrder(ctx, o)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &PurchaseResult{
		OrderID:       oid,
		ProductName:   p.Name,
		PriceUSD:      p.PriceUSD.StringFixed(2),
		ActivationKey: activationKey,
		QRData:        qr,
		Status:        "completed",
		ProviderRef:   providerRef,
	}, nil
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

func (uc *UseCase) ESIMOrder(ctx context.Context, planID string) (*domain.ESIMOrderResult, error) {
	if uc.esim == nil {
		return nil, wrapper.Wrap(domain.NewInvalidInput("PROVIDER_ERROR"))
	}

	return uc.esim.OrderESIM(ctx, planID)
}
