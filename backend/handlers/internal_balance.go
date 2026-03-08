package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// GetWalletHandler — GET /api/v1/user/wallet
// Возвращает текущий баланс Кошелька пользователя
func GetWalletHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ib, err := repository.GetInternalBalance(userID)
	if err != nil {
		http.Error(w, "Failed to get wallet: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ib)
}

// TopUpWalletHandler — POST /api/v1/user/wallet/topup
// Пополняет Кошелёк из balance_rub пользователя
func TopUpWalletHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.TopUpWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	ib, err := repository.TopUpInternalBalance(userID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ib)
}

// SetAutoTopupHandler — PATCH /api/v1/user/wallet/auto-topup
// Включает/выключает автопополнение карт из Кошелька
func SetAutoTopupHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if repository.GlobalDB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	_, err := repository.GlobalDB.Exec(
		`INSERT INTO internal_balances (user_id, auto_topup_enabled) VALUES ($1, $2)
		 ON CONFLICT (user_id) DO UPDATE SET auto_topup_enabled = $2, updated_at = NOW()`,
		userID, req.Enabled,
	)
	if err != nil {
		http.Error(w, "Failed to update auto-topup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"enabled": req.Enabled,
	})
}

// SetSpendingLimitHandler — PATCH /api/v1/user/cards/{id}/spending-limit
// Устанавливает лимит списания карты из Кошелька
func SetSpendingLimitHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	cardID, err := strconv.Atoi(vars["id"])
	if err != nil || cardID <= 0 {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	var req models.SpendingLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SpendingLimit.LessThan(decimal.Zero) {
		http.Error(w, "Spending limit cannot be negative", http.StatusBadRequest)
		return
	}

	if repository.GlobalDB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	// Проверяем принадлежность карты пользователю
	var ownerID int
	err = repository.GlobalDB.QueryRow("SELECT user_id FROM cards WHERE id = $1", cardID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}
	if ownerID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Обновляем spending_limit
	_, err = repository.GlobalDB.Exec(
		"UPDATE cards SET spending_limit = $1 WHERE id = $2",
		req.SpendingLimit, cardID,
	)
	if err != nil {
		http.Error(w, "Failed to update spending limit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "ok",
		"card_id":        cardID,
		"spending_limit": req.SpendingLimit,
	})
}
