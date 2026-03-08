package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
)

// GetReferralInfoHandler — GET /api/v1/user/referrals/info
// Возвращает реферальный код, ссылку, статистику и последних рефералов.
func GetReferralInfoHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	code, err := repository.GetUserReferralCode(userID)
	if err != nil {
		log.Printf("Failed to get referral code for user %d: %v", userID, err)
		http.Error(w, "Failed to get referral info", http.StatusInternalServerError)
		return
	}

	stats, err := repository.GetReferralStats(userID)
	if err != nil {
		log.Printf("Failed to get referral stats for user %d: %v", userID, err)
		http.Error(w, "Failed to get referral stats", http.StatusInternalServerError)
		return
	}

	recent, err := repository.GetRecentReferralsV2(userID)
	if err != nil {
		log.Printf("Failed to get recent referrals for user %d: %v", userID, err)
		recent = []repository.RecentReferral{}
	}

	domain := os.Getenv("APP_DOMAIN")
	if domain == "" {
		domain = "https://xplr.io"
	}
	referralLink := domain + "/register?ref=" + code

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"referral_code":       code,
		"referral_link":       referralLink,
		"stats":               stats,
		"recent_referrals":    recent,
		"reward_per_referral": 10,
		"bonus_for_new":       5,
		"commission_percent":  5.0,
	})
}
