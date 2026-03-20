package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
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

	amount := decimal.NewFromFloat(req.Amount)
	if err := repository.ProcessDeposit(userID, amount); err != nil {
		http.Error(w, "Deposit failed", http.StatusInternalServerError)
		return
	}

	log.Printf("[EVENT] User %d performed deposit (amount=%s ₽). Triggering notifications...", userID, amount.StringFixed(2))
	go service.NotifyUser(userID, "Пополнение баланса",
		fmt.Sprintf("💰 <b>Баланс пополнен</b>\n\n"+
			"Сумма: <b>%s ₽</b>\n\n"+
			"<a href=\"https://xplr.pro/wallet\">Открыть кошелёк</a>",
			amount.StringFixed(2)))

	user, _ := repository.GetUserByID(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"new_balance": user.BalanceRub.String()})
}
