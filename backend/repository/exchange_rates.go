package repository

import (
	"fmt"
	"log"

	"github.com/shopspring/decimal"
)

// ExchangeRate represents a currency pair rate with admin markup.
type ExchangeRate struct {
	ID             int    `json:"id"`
	CurrencyFrom   string `json:"currency_from"`
	CurrencyTo     string `json:"currency_to"`
	BaseRate       string `json:"base_rate"`
	MarkupPercent  string `json:"markup_percent"`
	FinalRate      string `json:"final_rate"`
	UpdatedAt      string `json:"updated_at"`
}

// GetAllExchangeRates returns all exchange rates.
func GetAllExchangeRates() ([]ExchangeRate, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	rows, err := GlobalDB.Query(
		`SELECT id, currency_from, currency_to, base_rate, markup_percent, final_rate, updated_at
		 FROM exchange_rates ORDER BY id`)
	if err != nil {
		log.Printf("ExchangeRates: query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var rates []ExchangeRate
	for rows.Next() {
		var r ExchangeRate
		var baseRate, markup, finalRate decimal.Decimal
		var updatedAt interface{}
		if err := rows.Scan(&r.ID, &r.CurrencyFrom, &r.CurrencyTo, &baseRate, &markup, &finalRate, &updatedAt); err != nil {
			log.Printf("ExchangeRates: scan error: %v", err)
			continue
		}
		r.BaseRate = baseRate.StringFixed(4)
		r.MarkupPercent = markup.StringFixed(2)
		r.FinalRate = finalRate.StringFixed(4)
		r.UpdatedAt = fmt.Sprintf("%v", updatedAt)
		rates = append(rates, r)
	}
	if rates == nil {
		rates = []ExchangeRate{}
	}
	return rates, nil
}

// GetFinalRate returns the final_rate for a given currency pair.
func GetFinalRate(from, to string) (decimal.Decimal, error) {
	if GlobalDB == nil {
		return decimal.Zero, fmt.Errorf("database connection not initialized")
	}
	var rate decimal.Decimal
	err := GlobalDB.QueryRow(
		"SELECT final_rate FROM exchange_rates WHERE currency_from = $1 AND currency_to = $2",
		from, to,
	).Scan(&rate)
	if err != nil {
		return decimal.Zero, fmt.Errorf("exchange rate %s/%s not found", from, to)
	}
	return rate, nil
}

// UpdateMarkupPercent updates the admin markup and recalculates final_rate.
func UpdateMarkupPercent(id int, newMarkup decimal.Decimal) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	// Recalculate: final_rate = base_rate * (1 + markup_percent / 100)
	_, err := GlobalDB.Exec(
		`UPDATE exchange_rates 
		 SET markup_percent = $1, 
		     final_rate = base_rate * (1 + $1 / 100),
		     updated_at = NOW()
		 WHERE id = $2`,
		newMarkup, id,
	)
	if err != nil {
		log.Printf("UpdateMarkupPercent: DB error: %v", err)
		return fmt.Errorf("failed to update markup")
	}
	log.Printf("✅ Exchange rate id=%d markup updated to %s%%", id, newMarkup.String())
	return nil
}

// UpdateBaseRate updates the base_rate and recalculates final_rate (used by rate fetcher).
func UpdateBaseRate(from, to string, newBase decimal.Decimal) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE exchange_rates 
		 SET base_rate = $1, 
		     final_rate = $1 * (1 + markup_percent / 100),
		     updated_at = NOW()
		 WHERE currency_from = $2 AND currency_to = $3`,
		newBase, from, to,
	)
	if err != nil {
		log.Printf("UpdateBaseRate %s/%s: DB error: %v", from, to, err)
		return fmt.Errorf("failed to update base rate")
	}
	log.Printf("✅ Exchange rate %s/%s base updated to %s", from, to, newBase.String())
	return nil
}

// SeedDefaultExchangeRates inserts default rates if table is empty.
func SeedDefaultExchangeRates() {
	if GlobalDB == nil {
		return
	}
	var count int
	GlobalDB.QueryRow("SELECT COUNT(*) FROM exchange_rates").Scan(&count)
	if count > 0 {
		return
	}

	defaults := []struct {
		From, To string
		Base     string
		Markup   string
	}{
		{"RUB", "USD", "96.5000", "3.00"},
		{"RUB", "EUR", "104.0000", "3.00"},
	}

	for _, d := range defaults {
		base, _ := decimal.NewFromString(d.Base)
		markup, _ := decimal.NewFromString(d.Markup)
		hundred := decimal.NewFromInt(100)
		finalRate := base.Mul(decimal.NewFromInt(1).Add(markup.Div(hundred)))

		_, err := GlobalDB.Exec(
			`INSERT INTO exchange_rates (currency_from, currency_to, base_rate, markup_percent, final_rate)
			 VALUES ($1, $2, $3, $4, $5)`,
			d.From, d.To, base, markup, finalRate,
		)
		if err != nil {
			log.Printf("Warning: failed to seed exchange rate %s/%s: %v", d.From, d.To, err)
		} else {
			log.Printf("✅ Seeded exchange rate %s/%s: base=%s, markup=%s%%, final=%s",
				d.From, d.To, base.String(), markup.String(), finalRate.StringFixed(4))
		}
	}
}
