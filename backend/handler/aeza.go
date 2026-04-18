package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/djalben/xplr-core/backend/service"
)

// GetAezaBalanceHandler - GET /api/v1/admin/infra/balance
// Returns the current Aeza hosting balance (admin only).
func GetAezaBalanceHandler(w http.ResponseWriter, r *http.Request) {
	balance, err := service.GetAezaBalance()
	if err != nil {
		log.Printf("[AEZA-HANDLER] ❌ %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  err.Error(),
			"status": "unavailable",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// CheckAezaBalanceHandler - POST /api/v1/admin/infra/balance/check
// Triggers a manual balance check + notification if low (admin only).
func CheckAezaBalanceHandler(w http.ResponseWriter, r *http.Request) {
	service.CheckAezaBalanceAndNotify()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "checked",
	})
}
