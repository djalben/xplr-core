package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
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

	// Check if SBP is enabled
	var sbpEnabled string
	err := GlobalDB.QueryRow(`SELECT setting_value FROM system_settings WHERE setting_key = 'sbp_enabled'`).Scan(&sbpEnabled)
	if err == nil && sbpEnabled != "true" {
		http.Error(w, "СБП временно недоступен. Попробуйте позже.", http.StatusServiceUnavailable)
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

	log.Printf("[EVENT] User %d performed wallet_topup (amount=%s RUB, new_balance=$%s). Triggering notifications...", userID, req.Amount.StringFixed(0), ib.MasterBalance.StringFixed(2))

	// Notify user about successful topup
	go func() {
		service.NotifyUser(userID, "Кошелёк пополнен",
			fmt.Sprintf("💰 <b>Кошелёк пополнен</b>\n\n"+
				"Сумма: <b>%s ₽</b>\n"+
				"Баланс: <b>$%s</b>\n\n"+
				"<a href=\"https://xplr.pro/wallet\">Открыть кошелёк</a>",
				req.Amount.StringFixed(0), ib.MasterBalance.StringFixed(2)))
		log.Printf("[NOTIFY] Message sent to UserID: %d via NotifyUser (wallet topup %s RUB)", userID, req.Amount.StringFixed(0))
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ib)
}

// TransferWalletToCardHandler — POST /api/v1/user/wallet/transfer-to-card
// Переводит средства из Кошелька (USD) на баланс карты. Если карта в EUR, конвертирует.
func TransferWalletToCardHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CardID   json.Number `json:"card_id"`
		Amount   float64     `json:"amount"`
		Currency string      `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cardIDStr := req.CardID.String()
	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil || cardID <= 0 {
		http.Error(w, "Invalid card_id", http.StatusBadRequest)
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	if amount.LessThanOrEqual(decimal.Zero) {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	ib, err := repository.TransferWalletToCard(userID, cardID, amount, req.Currency)
	if err != nil {
		if err.Error() == "нет доступа к этой карте" || err.Error() == "карта не найдена" {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else if err.Error() == "кошелёк не найден — пополните баланс" {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	log.Printf("[EVENT] User %d performed fund_card (card=%d, amount=%s %s). Triggering notifications...", userID, cardID, amount.StringFixed(2), req.Currency)

	// Notify user about card funding
	go func() {
		var cardLast4 string
		if card, err := repository.GetCardByID(cardID); err == nil {
			cardLast4 = card.Last4Digits
		}
		curr := req.Currency
		if curr == "" {
			curr = "USD"
		}
		service.NotifyUser(userID, "Карта пополнена",
			fmt.Sprintf("💳 <b>Карта пополнена</b>\n\n"+
				"Карта: *%s\n"+
				"Сумма: <b>%s %s</b>\n\n"+
				"<a href=\"https://xplr.pro/cards\">Открыть карты</a>",
				cardLast4, amount.StringFixed(2), curr))
	}()

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
