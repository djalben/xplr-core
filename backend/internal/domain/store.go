package domain

import "time"

type StoreProductType string

const (
	StoreProductTypeESIM    StoreProductType = "esim"
	StoreProductTypeDigital StoreProductType = "digital"
	StoreProductTypeVPN     StoreProductType = "vpn"
)

type StoreOrderStatus string

const (
	StoreOrderStatusPending   StoreOrderStatus = "PENDING"
	StoreOrderStatusReady     StoreOrderStatus = "READY"
	StoreOrderStatusFailed    StoreOrderStatus = "FAILED"
	StoreOrderStatusCompleted StoreOrderStatus = "COMPLETED"
	StoreOrderStatusDeleted   StoreOrderStatus = "DELETED"
)

type StoreCategory struct {
	ID          UUID      `json:"id" db:"id"`
	Slug        string    `json:"slug" db:"slug"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Icon        string    `json:"icon" db:"icon"`
	ImageURL    string    `json:"imageUrl" db:"image_url"`
	SortOrder   int       `json:"sortOrder" db:"sort_order"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

type StoreProduct struct {
	ID           UUID             `json:"id" db:"id"`
	CategoryID   UUID             `json:"categoryId" db:"category_id"`
	CategorySlug string           `json:"categorySlug" db:"category_slug"`
	Provider     string           `json:"provider" db:"provider"`
	ExternalID   string           `json:"externalId" db:"external_id"`
	Name         string           `json:"name" db:"name"`
	Description  string           `json:"description" db:"description"`
	Country      string           `json:"country" db:"country"`
	CountryCode  string           `json:"countryCode" db:"country_code"`
	ProductType  StoreProductType `json:"productType" db:"product_type"`
	PriceUSD     Numeric          `json:"priceUsd" db:"price_usd"`
	CostPrice    Numeric          `json:"costPrice" db:"cost_price"`
	MarkupPct    Numeric          `json:"markupPercent" db:"markup_percent"`
	OldPrice     Numeric          `json:"oldPrice" db:"old_price"`
	DataGB       string           `json:"dataGb" db:"data_gb"`
	ValidityDays int              `json:"validityDays" db:"validity_days"`
	ImageURL     string           `json:"imageUrl" db:"image_url"`
	InStock      bool             `json:"inStock" db:"in_stock"`
	Meta         string           `json:"meta" db:"meta"`
	SortOrder    int              `json:"sortOrder" db:"sort_order"`
}

type StoreOrder struct {
	ID            UUID             `json:"id" db:"id"`
	UserID        UUID             `json:"userId" db:"user_id"`
	ProductID     *UUID            `json:"productId,omitempty" db:"product_id"`
	ProductName   string           `json:"productName" db:"product_name"`
	PriceUSD      Numeric          `json:"priceUsd" db:"price_usd"`
	Status        StoreOrderStatus `json:"status" db:"status"`
	ActivationKey string           `json:"activationKey" db:"activation_key"`
	QRData        string           `json:"qrData" db:"qr_data"`
	ProviderRef   string           `json:"providerRef" db:"provider_ref"`
	Meta          string           `json:"meta" db:"meta"`
	CreatedAt     time.Time        `json:"createdAt" db:"created_at"`
}
