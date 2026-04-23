package domain

import "time"

type AdminVPNOrderRow struct {
	UserEmail     string    `json:"email" db:"email"`
	ProductName   string    `json:"productName" db:"product_name"`
	PriceUSD      float64   `json:"priceUsd" db:"price_usd"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	ProviderRef   string    `json:"providerRef" db:"provider_ref"`
	ActivationKey string    `json:"activationKey" db:"activation_key"`
	Meta          string    `json:"meta" db:"meta"`
}
