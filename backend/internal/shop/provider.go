package shop

import "github.com/shopspring/decimal"

// ══════════════════════════════════════════════════════════════
// ProductProvider — universal interface for any dropshipping supplier.
// Every external supplier (MobiMatter, Razer Gold, etc.) implements this.
// ══════════════════════════════════════════════════════════════

// CatalogProduct represents a single product returned by the supplier.
type CatalogProduct struct {
	ExternalID  string          `json:"external_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`    // "esim", "digital", "gift_card"
	Country     string          `json:"country"`     // country name (for eSIM)
	CountryCode string          `json:"country_code"` // ISO-2 (for eSIM)
	CostPrice   decimal.Decimal `json:"cost_price"`  // supplier price in USD
	Currency    string          `json:"currency"`    // typically "USD"
	ImageURL    string          `json:"image_url"`
	InStock     bool            `json:"in_stock"`
	Meta        map[string]any  `json:"meta,omitempty"` // provider-specific extra fields
}

// OrderResult is returned after placing an order with the supplier.
type OrderResult struct {
	ProviderRef   string `json:"provider_ref"`   // supplier's order/reference ID
	ActivationKey string `json:"activation_key"`  // digital product key (if applicable)
	QRData        string `json:"qr_data"`         // QR / LPA string (for eSIM)
	ICCID         string `json:"iccid"`           // eSIM ICCID (if applicable)
	Status        string `json:"status"`          // "completed", "pending", "failed"
	RawResponse   []byte `json:"raw_response,omitempty"`
}

// OrderStatus represents the current state of a supplier order.
type OrderStatus struct {
	ProviderRef   string `json:"provider_ref"`
	Status        string `json:"status"` // "pending", "completed", "failed", "refunded"
	ActivationKey string `json:"activation_key,omitempty"`
	QRData        string `json:"qr_data,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// BalanceInfo holds the supplier deposit/balance information.
type BalanceInfo struct {
	BalanceUSD decimal.Decimal `json:"balance_usd"`
	Currency   string          `json:"currency"`
}

// ProductProvider — every external supplier must implement this interface.
type ProductProvider interface {
	// Name returns the unique provider identifier (e.g. "mobimatter", "razer").
	Name() string

	// GetCatalog fetches the full product catalog from the supplier.
	GetCatalog() ([]CatalogProduct, error)

	// CreateOrder places an order for a product by its external ID.
	// Returns activation data (key / QR) on success.
	CreateOrder(externalProductID string) (*OrderResult, error)

	// CheckStatus checks the current status of a previously placed order.
	CheckStatus(providerRef string) (*OrderStatus, error)

	// GetBalance returns the current deposit balance at the supplier.
	// Returns nil, nil if the provider does not support balance queries.
	GetBalance() (*BalanceInfo, error)
}
