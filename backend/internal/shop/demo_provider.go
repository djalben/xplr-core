package shop

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// DemoProvider is a test/fallback provider that returns simulated data.
// Used when no real supplier API key is configured.
type DemoProvider struct{}

func NewDemoProvider() *DemoProvider { return &DemoProvider{} }

func (d *DemoProvider) Name() string { return "demo" }

func (d *DemoProvider) GetCatalog() ([]CatalogProduct, error) {
	return []CatalogProduct{
		{ExternalID: "demo-steam-50", Name: "Steam $50", Description: "Подарочная карта Steam", Category: "digital", CostPrice: decimal.NewFromFloat(42.00), Currency: "USD", InStock: true},
		{ExternalID: "demo-spotify-1m", Name: "Spotify Premium 1 мес", Description: "Подписка Spotify Premium", Category: "digital", CostPrice: decimal.NewFromFloat(8.50), Currency: "USD", InStock: true},
		{ExternalID: "demo-netflix-1m", Name: "Netflix Standard 1 мес", Description: "Подписка Netflix Standard", Category: "digital", CostPrice: decimal.NewFromFloat(12.00), Currency: "USD", InStock: true},
	}, nil
}

func (d *DemoProvider) CreateOrder(externalProductID string) (*OrderResult, error) {
	ref := fmt.Sprintf("DEMO-%s-%d", externalProductID, time.Now().UnixMilli())
	key := fmt.Sprintf("XPLR-%s-%d", externalProductID, time.Now().UnixMilli())
	return &OrderResult{
		ProviderRef:   ref,
		ActivationKey: key,
		Status:        "completed",
	}, nil
}

func (d *DemoProvider) CheckStatus(providerRef string) (*OrderStatus, error) {
	return &OrderStatus{
		ProviderRef: providerRef,
		Status:      "completed",
	}, nil
}

func (d *DemoProvider) GetBalance() (*BalanceInfo, error) {
	return &BalanceInfo{
		BalanceUSD: decimal.NewFromFloat(999.99),
		Currency:   "USD",
	}, nil
}
