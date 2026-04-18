package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/djalben/xplr-core/backend/providers/vless"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/djalben/xplr-core/backend/shop"
)

// GetAezaBalanceHandler - GET /api/v1/admin/infra/balance
// Returns the current Aeza hosting balance (admin only).
// On persistent 5xx from Aeza, returns status:"maintenance" (200 OK) instead of an error.
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

	// Maintenance status is returned as 200 so frontend can handle it gracefully
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

// GetActiveVPNKeysHandler - GET /api/v1/admin/infra/active-keys
// Returns the count of active VPN keys from the 3X-UI panel.
func GetActiveVPNKeysHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	provider := shop.GetRegistry().Get("vless")
	if provider == nil {
		json.NewEncoder(w).Encode(map[string]any{"active_keys": 0, "error": "vless provider not registered"})
		return
	}

	vp, ok := provider.(*vless.VlessProvider)
	if !ok {
		json.NewEncoder(w).Encode(map[string]any{"active_keys": 0, "error": "provider type assertion failed"})
		return
	}

	count, err := vp.GetActiveClients()
	if err != nil {
		log.Printf("[INFRA] ❌ GetActiveClients error: %v", err)
		json.NewEncoder(w).Encode(map[string]any{"active_keys": 0, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"active_keys": count})
}
