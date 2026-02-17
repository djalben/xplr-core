package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/djalben/xplr-core/backend/repository"
	"github.com/shopspring/decimal"
)

// StartExchangeRateFetcher starts a background goroutine that updates base rates every hour.
func StartExchangeRateFetcher() {
	log.Println("[EXCHANGE] Rate fetcher started (updating every 1 hour)")

	// Immediate first fetch
	fetchAndUpdateRates()

	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		fetchAndUpdateRates()
	}
}

func fetchAndUpdateRates() {
	log.Println("[EXCHANGE] Fetching latest exchange rates...")

	// Try free API: exchangerate-api.com (no key needed for USD base)
	rubUsd, err := fetchRateFromAPI("USD", "RUB")
	if err != nil {
		log.Printf("[EXCHANGE] Warning: API fetch failed, using fallback: %v", err)
		return
	}

	if rubUsd.GreaterThan(decimal.Zero) {
		if err := repository.UpdateBaseRate("RUB", "USD", rubUsd); err != nil {
			log.Printf("[EXCHANGE] Failed to update RUB/USD: %v", err)
		}
	}

	// EUR rate: derive from USD/EUR cross rate
	eurUsd, err := fetchRateFromAPI("USD", "EUR")
	if err == nil && eurUsd.GreaterThan(decimal.Zero) {
		// RUB/EUR = RUB/USD * USD/EUR  (how many RUB per 1 EUR)
		// Actually: if rubUsd = how many RUB per 1 USD, and eurUsd = how many EUR per 1 USD
		// then RUB per 1 EUR = rubUsd / eurUsd
		rubEur := rubUsd.Div(eurUsd)
		if err := repository.UpdateBaseRate("RUB", "EUR", rubEur); err != nil {
			log.Printf("[EXCHANGE] Failed to update RUB/EUR: %v", err)
		}
	}

	log.Println("[EXCHANGE] Rate update complete")
}

// fetchRateFromAPI fetches the exchange rate for base->target from a free API.
// Returns how many units of `target` per 1 unit of `base`.
func fetchRateFromAPI(base, target string) (decimal.Decimal, error) {
	url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", base)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return decimal.Zero, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return decimal.Zero, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Result string                 `json:"result"`
		Rates  map[string]json.Number `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return decimal.Zero, fmt.Errorf("JSON decode error: %v", err)
	}

	if result.Result != "success" {
		return decimal.Zero, fmt.Errorf("API result: %s", result.Result)
	}

	rateStr, ok := result.Rates[target]
	if !ok {
		return decimal.Zero, fmt.Errorf("target currency %s not found in response", target)
	}

	rate, err := decimal.NewFromString(rateStr.String())
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse rate: %v", err)
	}

	log.Printf("[EXCHANGE] Fetched %s/%s = %s", base, target, rate.StringFixed(4))
	return rate, nil
}
