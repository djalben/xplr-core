package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
)

// GetDashboardStatsHandler — GET /api/v1/user/dashboard-stats
// Returns aggregated dashboard data: today total, 5 recent transactions, 7-day chart.
func GetDashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := repository.GetDashboardStats(userID)
	if err != nil {
		log.Printf("[DASHBOARD-STATS] Error for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
