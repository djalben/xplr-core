package providers

// ══════════════════════════════════════════════════════════════
// eSIM domain types + provider interface.
// The ONLY live source of truth is the Keepgo (Esimba) API v2.
// No demo/fallback data is used anywhere in the storefront.
// ══════════════════════════════════════════════════════════════

// ESIMDestination — a country/region where eSIM plans are available.
type ESIMDestination struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	FlagEmoji   string `json:"flag_emoji"`
	PlanCount   int    `json:"plan_count"`
}

// ESIMPlan — a single eSIM plan (data bundle). All prices are in USD.
type ESIMPlan struct {
	PlanID       string  `json:"plan_id"`
	Provider     string  `json:"provider"`
	Name         string  `json:"name"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	DataGB       string  `json:"data_gb"`
	ValidityDays int     `json:"validity_days"`
	PriceUSD     float64 `json:"price_usd"`
	OldPrice     float64 `json:"old_price"`
	CostPrice    float64 `json:"cost_price,omitempty"`
	Description  string  `json:"description"`
	InStock      bool    `json:"in_stock"`
}

// ESIMOrderResult — result of ordering an eSIM.
type ESIMOrderResult struct {
	OrderID     string `json:"order_id"`
	QRData      string `json:"qr_data"`
	LPA         string `json:"lpa"`
	SMDP        string `json:"smdp"`
	MatchingID  string `json:"matching_id"`
	ICCID       string `json:"iccid"`
	ProviderRef string `json:"provider_ref"`
}

// ESIMProvider interface — any eSIM provider must implement these.
type ESIMProvider interface {
	GetDestinations() ([]ESIMDestination, error)
	GetPlans(countryCode string) ([]ESIMPlan, error)
	OrderESIM(planID string) (*ESIMOrderResult, error)
	CheckAvailability(planID string) (bool, error)
	Name() string
}

// ══════════════════════════════════════════════════════════════
// Singleton accessor
// ══════════════════════════════════════════════════════════════

// GetESIMProvider returns the singleton eSIM provider (Esimba/Keepgo API v2).
func GetESIMProvider() ESIMProvider {
	return getEsimbaProvider()
}

// ══════════════════════════════════════════════════════════════
// Utility
// ══════════════════════════════════════════════════════════════

// countryFlag converts an ISO-3166-1 alpha-2 code into a flag emoji.
func countryFlag(code string) string {
	if len(code) < 2 {
		return "🌍"
	}
	code = code[:2]
	runes := []rune(code)
	return string([]rune{runes[0] - 'A' + 0x1F1E6, runes[1] - 'A' + 0x1F1E6})
}
