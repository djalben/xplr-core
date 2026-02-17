package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aalabin/xplr/middleware"
	"github.com/aalabin/xplr/repository"
	"github.com/shopspring/decimal"
)

type DepositRequest struct {
	Amount float64 `json:"amount"`
}

func ProcessDepositHandler(w http.ResponseWriter, r *http.Request) {
	// Синхронизировано: используем UserIDKey
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if err := repository.ProcessDeposit(userID, decimal.NewFromFloat(req.Amount)); err != nil {
		http.Error(w, "Deposit failed", http.StatusInternalServerError)
		return
	}

	user, _ := repository.GetUserByID(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"new_balance": user.BalanceRub.String()})
}