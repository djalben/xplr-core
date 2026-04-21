package domain

type ESIMDestination struct {
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName"`
	FlagEmoji   string `json:"flagEmoji"`
	PlanCount   int    `json:"planCount"`
}

type ESIMPlan struct {
	PlanID       string  `json:"planId"`
	Provider     string  `json:"provider"`
	Name         string  `json:"name"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"countryCode"`
	DataGB       string  `json:"dataGb"`
	ValidityDays int     `json:"validityDays"`
	PriceUSD     Numeric `json:"priceUsd"`
	OldPrice     Numeric `json:"oldPrice"`
	CostPrice    Numeric `json:"costPrice"`
	Description  string  `json:"description"`
	InStock      bool    `json:"inStock"`
}

type ESIMOrderResult struct {
	OrderID     string `json:"orderId"`
	QRData      string `json:"qrData"`
	LPA         string `json:"lpa"`
	SMDP        string `json:"smdp"`
	MatchingID  string `json:"matchingId"`
	ICCID       string `json:"iccid"`
	ProviderRef string `json:"providerRef"`
}
