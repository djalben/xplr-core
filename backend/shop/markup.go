package shop

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// ══════════════════════════════════════════════════════════════
// Global Markup — reads global_markup_percent from system_settings.
// Caches for 60 seconds to avoid per-request DB queries.
// ══════════════════════════════════════════════════════════════

var (
	cachedMarkup     decimal.Decimal
	cachedMarkupAt   time.Time
	cachedMarkupOnce sync.Once
	markupMu         sync.Mutex
	markupDB         *sql.DB
)

const markupCacheTTL = 60 * time.Second
const defaultMarkupPercent = 20.0

// InitMarkup sets the DB reference for markup lookups.
func InitMarkup(db *sql.DB) {
	markupDB = db
}

// GetGlobalMarkup returns the current global_markup_percent from DB (cached).
func GetGlobalMarkup() decimal.Decimal {
	markupMu.Lock()
	defer markupMu.Unlock()

	if time.Since(cachedMarkupAt) < markupCacheTTL {
		return cachedMarkup
	}

	// Refresh from DB
	cachedMarkupOnce.Do(func() {
		cachedMarkup = decimal.NewFromFloat(defaultMarkupPercent)
	})

	if markupDB == nil {
		return cachedMarkup
	}

	var val string
	err := markupDB.QueryRow(
		`SELECT setting_value FROM system_settings WHERE setting_key = 'global_markup_percent'`,
	).Scan(&val)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[SHOP-MARKUP] ⚠️ Failed to read global_markup_percent: %v", err)
		}
		// Keep previous cached value or default
		cachedMarkupAt = time.Now()
		return cachedMarkup
	}

	parsed, err := decimal.NewFromString(val)
	if err != nil {
		log.Printf("[SHOP-MARKUP] ⚠️ Invalid global_markup_percent value %q: %v", val, err)
		cachedMarkupAt = time.Now()
		return cachedMarkup
	}

	cachedMarkup = parsed
	cachedMarkupAt = time.Now()
	log.Printf("[SHOP-MARKUP] Global markup refreshed: %s%%", cachedMarkup.StringFixed(1))
	return cachedMarkup
}

// ApplyMarkup calculates the retail price: cost * (1 + markup/100), rounded to .90.
func ApplyMarkup(costPrice, markupPercent decimal.Decimal) decimal.Decimal {
	if costPrice.IsZero() {
		return decimal.Zero
	}
	multiplier := decimal.NewFromInt(1).Add(markupPercent.Div(decimal.NewFromInt(100)))
	raw := costPrice.Mul(multiplier)

	// Round up to nearest .90
	floor := raw.Floor()
	frac := raw.Sub(floor)
	threshold := decimal.NewFromFloat(0.90)
	if frac.LessThanOrEqual(threshold) {
		return floor.Add(threshold)
	}
	return floor.Add(decimal.NewFromInt(1)).Add(threshold)
}

// ApplyGlobalMarkup applies the cached global markup to a cost price.
func ApplyGlobalMarkup(costPrice decimal.Decimal) decimal.Decimal {
	return ApplyMarkup(costPrice, GetGlobalMarkup())
}
