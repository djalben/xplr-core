package domain

import "time"

type ExchangeRate struct {
	ID            UUID      `json:"id" db:"id"`
	CurrencyFrom  string    `json:"currencyFrom" db:"currency_from"`
	CurrencyTo    string    `json:"currencyTo" db:"currency_to"`
	BaseRate      Numeric   `json:"baseRate" db:"base_rate"`
	MarkupPercent Numeric   `json:"markupPercent" db:"markup_percent"`
	FinalRate     Numeric   `json:"finalRate" db:"final_rate"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}
