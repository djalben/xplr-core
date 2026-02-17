package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
)

// GetReferralStatsHandler - Возвращает статистику реферальной программы пользователя
func GetReferralStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Извлечение UserID из контекста
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Получение статистики реферальной программы
	stats, err := repository.GetReferralStats(userID)
	if err != nil {
		log.Printf("Error fetching referral stats for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch referral stats", http.StatusInternalServerError)
		return
	}

	// Формирование ответа
	response := struct {
		TotalReferrals  int    `json:"total_referrals"`
		ActiveReferrals int    `json:"active_referrals"`
		TotalCommission string `json:"total_commission"`
		ReferralCode    string `json:"referral_code"`
	}{
		TotalReferrals:  stats.TotalReferrals,
		ActiveReferrals: stats.ActiveReferrals,
		TotalCommission: stats.TotalCommission.String(),
		ReferralCode:    stats.ReferralCode,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding referral stats response for user %d: %v", userID, err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetReferralListHandler - GET /api/v1/user/referrals/list
func GetReferralListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	list, err := repository.GetReferralList(userID)
	if err != nil {
		log.Printf("Error fetching referral list for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch referral list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
