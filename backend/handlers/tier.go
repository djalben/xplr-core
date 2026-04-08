package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/shopspring/decimal"
)

// UpgradeTierHandler - POST /api/v1/user/upgrade-tier
// Upgrades user to Gold tier, deducts from wallet, sets expiration
func UpgradeTierHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Gold tier price from system_settings
	var goldPriceStr string
	err := GlobalDB.QueryRow(`SELECT setting_value FROM system_settings WHERE setting_key = 'gold_tier_price'`).Scan(&goldPriceStr)
	if err != nil {
		log.Printf("[TIER-UPGRADE] Error fetching gold_tier_price: %v", err)
		goldPriceStr = "50.00" // fallback
	}
	goldPrice, _ := decimal.NewFromString(goldPriceStr)

	// Get duration
	var durationDaysStr string
	err = GlobalDB.QueryRow(`SELECT setting_value FROM system_settings WHERE setting_key = 'gold_tier_duration_days'`).Scan(&durationDaysStr)
	if err != nil {
		log.Printf("[TIER-UPGRADE] Error fetching gold_tier_duration_days: %v", err)
		durationDaysStr = "30" // fallback
	}
	durationDays, _ := strconv.Atoi(durationDaysStr)

	// Check current tier and existing expiry
	var currentTier string
	var tierExpiresAt sql.NullTime
	err = GlobalDB.QueryRow(`SELECT COALESCE(tier, 'standard'), tier_expires_at FROM users WHERE id = $1`, userID).Scan(&currentTier, &tierExpiresAt)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check wallet balance
	wallet, err := repository.GetInternalBalance(userID)
	if err != nil {
		http.Error(w, "Failed to get wallet balance", http.StatusInternalServerError)
		return
	}

	if wallet.MasterBalance.LessThan(goldPrice) {
		http.Error(w, "Insufficient wallet balance", http.StatusPaymentRequired)
		return
	}

	// Deduct from wallet
	details := "Gold tier upgrade — $" + goldPrice.StringFixed(2) + " за " + durationDaysStr + " дн."
	err = repository.DeductWalletBalance(userID, goldPrice, details)
	if err != nil {
		log.Printf("[TIER-UPGRADE] Failed to deduct wallet: %v", err)
		http.Error(w, "Failed to process payment", http.StatusInternalServerError)
		return
	}

	// Cumulative expiry: max(current_expires, now) + duration_days
	baseTime := time.Now()
	if tierExpiresAt.Valid && tierExpiresAt.Time.After(baseTime) {
		baseTime = tierExpiresAt.Time
	}
	expiresAt := baseTime.Add(time.Duration(durationDays) * 24 * time.Hour)

	_, err = GlobalDB.Exec(`UPDATE users SET tier = 'gold', tier_expires_at = $1 WHERE id = $2`, expiresAt, userID)
	if err != nil {
		log.Printf("[TIER-UPGRADE] Failed to update tier: %v", err)
		// Refund wallet
		_, _ = repository.TopUpInternalBalance(userID, goldPrice)
		http.Error(w, "Failed to upgrade tier", http.StatusInternalServerError)
		return
	}

	// Log transaction
	goldPriceFloat, _ := goldPrice.Float64()
	_, err = GlobalDB.Exec(`
		INSERT INTO transactions (user_id, amount, fee, transaction_type, status, details, currency)
		VALUES ($1, $2, 0, 'tier_upgrade', 'completed', $3, 'USD')
	`, userID, goldPriceFloat, details)
	if err != nil {
		log.Printf("[TIER-UPGRADE] Failed to log transaction: %v", err)
	}

	isExtension := currentTier == "gold" && tierExpiresAt.Valid && tierExpiresAt.Time.After(time.Now())
	actionLabel := "активирован"
	if isExtension {
		actionLabel = "продлён"
	}

	// Notify user
	go service.NotifyUser(userID, "Gold "+actionLabel,
		fmt.Sprintf("🏆 <b>Gold %s!</b>\n\n"+
			"Срок: до <b>%s</b>\n"+
			"Лимит карт: <b>15</b>\n"+
			"Стоимость: <b>$%s</b>\n\n"+
			"<a href=\"https://xplr.pro/dashboard\">Открыть дашборд</a>",
			actionLabel, expiresAt.Format("02.01.2006"), goldPrice.StringFixed(2)))

	log.Printf("[TIER-UPGRADE] ✅ User %d Gold %s (expires: %s, paid: $%s)", userID, actionLabel, expiresAt.Format("2006-01-02"), goldPrice.StringFixed(2))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"tier":       "gold",
		"expires_at": expiresAt,
		"paid":       goldPrice,
	})
}

// GetTierInfoHandler - GET /api/v1/user/tier-info
// Returns current tier info and pricing
func GetTierInfoHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var tier string
	var tierExpiresAt sql.NullTime
	err := GlobalDB.QueryRow(`SELECT COALESCE(tier, 'standard'), tier_expires_at FROM users WHERE id = $1`, userID).Scan(&tier, &tierExpiresAt)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get pricing
	var goldPrice string
	var durationDays string
	GlobalDB.QueryRow(`SELECT setting_value FROM system_settings WHERE setting_key = 'gold_tier_price'`).Scan(&goldPrice)
	GlobalDB.QueryRow(`SELECT setting_value FROM system_settings WHERE setting_key = 'gold_tier_duration_days'`).Scan(&durationDays)

	// Count user's cards
	var cardCount int
	GlobalDB.QueryRow(`SELECT COUNT(*) FROM cards WHERE user_id = $1`, userID).Scan(&cardCount)

	// Determine limits
	cardLimit := 3
	if tier == "gold" && tierExpiresAt.Valid && tierExpiresAt.Time.After(time.Now()) {
		cardLimit = 15
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tier":            tier,
		"tier_expires_at": tierExpiresAt,
		"card_limit":      cardLimit,
		"current_cards":   cardCount,
		"gold_price":      goldPrice,
		"gold_duration":   durationDays,
		"can_issue_more":  cardCount < cardLimit,
	})
}
