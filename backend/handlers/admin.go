package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/djalben/xplr-core/backend/repository"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// AdminAdjustBalanceHandler - PATCH /api/v1/admin/users/{id}/balance
func AdminAdjustBalanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID, err := strconv.Atoi(vars["id"])
	if err != nil || targetID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	var req struct {
		Amount decimal.Decimal `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Amount.IsZero() {
		http.Error(w, "amount must not be zero", http.StatusBadRequest)
		return
	}
	newBal, err := repository.AdminAdjustBalance(targetID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":     targetID,
		"adjustment":  req.Amount.String(),
		"new_balance": newBal,
	})
}

// AdminToggleRoleHandler - PATCH /api/v1/admin/users/{id}/role
func AdminToggleRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID, err := strconv.Atoi(vars["id"])
	if err != nil || targetID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	newVal, err := repository.AdminToggleRole(targetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  targetID,
		"is_admin": newVal,
	})
}

// AdminSetUserStatusHandler - PATCH /api/v1/admin/users/{id}/status
func AdminSetUserStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID, err := strconv.Atoi(vars["id"])
	if err != nil || targetID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	status := strings.TrimSpace(strings.ToUpper(req.Status))
	if status != "ACTIVE" && status != "BANNED" {
		http.Error(w, "status must be ACTIVE or BANNED", http.StatusBadRequest)
		return
	}
	if err := repository.AdminSetUserStatus(targetID, status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": targetID,
		"status":  status,
	})
}

// AdminGetExchangeRatesHandler - GET /api/v1/admin/rates
func AdminGetExchangeRatesHandler(w http.ResponseWriter, r *http.Request) {
	rates, err := repository.GetAllExchangeRates()
	if err != nil {
		log.Printf("Error fetching exchange rates: %v", err)
		http.Error(w, "Failed to fetch exchange rates", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rates)
}

// AdminUpdateMarkupHandler - PATCH /api/v1/admin/rates/{id}/markup
func AdminUpdateMarkupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rateID, err := strconv.Atoi(vars["id"])
	if err != nil || rateID <= 0 {
		http.Error(w, "invalid rate id", http.StatusBadRequest)
		return
	}
	var req struct {
		MarkupPercent decimal.Decimal `json:"markup_percent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.MarkupPercent.LessThan(decimal.Zero) {
		http.Error(w, "markup_percent cannot be negative", http.StatusBadRequest)
		return
	}
	if err := repository.UpdateMarkupPercent(rateID, req.MarkupPercent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Return updated rates
	rates, _ := repository.GetAllExchangeRates()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rates)
}

// PublicGetExchangeRatesHandler - GET /api/v1/rates (no auth needed)
func PublicGetExchangeRatesHandler(w http.ResponseWriter, r *http.Request) {
	rates, err := repository.GetAllExchangeRates()
	if err != nil {
		log.Printf("Error fetching public exchange rates: %v", err)
		http.Error(w, "Failed to fetch rates", http.StatusInternalServerError)
		return
	}
	// Return only user-facing fields
	type PublicRate struct {
		CurrencyFrom string `json:"currency_from"`
		CurrencyTo   string `json:"currency_to"`
		FinalRate    string `json:"final_rate"`
		UpdatedAt    string `json:"updated_at"`
	}
	var publicRates []PublicRate
	for _, r := range rates {
		publicRates = append(publicRates, PublicRate{
			CurrencyFrom: r.CurrencyFrom,
			CurrencyTo:   r.CurrencyTo,
			FinalRate:    r.FinalRate,
			UpdatedAt:    r.UpdatedAt,
		})
	}
	if publicRates == nil {
		publicRates = []PublicRate{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publicRates)
}

// AdminStatsHandler - GET /api/v1/admin/stats
func AdminStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := repository.GetAdminStats()
	if err != nil {
		log.Printf("Error fetching admin stats: %v", err)
		http.Error(w, "Failed to fetch admin stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// AdminUsersHandler - GET /api/v1/admin/users
func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := repository.GetAllUsersForAdmin()
	if err != nil {
		log.Printf("Error fetching admin users: %v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetAdminTransactionReportHandler - Выдает отчет обо всех транзакциях платформы.
func GetAdminTransactionReportHandler(w http.ResponseWriter, r *http.Request) {
	// В этом обработчике мы уверены, что пользователь является Администратором

	report, err := repository.GetAdminTransactionReport()
	if err != nil {
		log.Printf("Error fetching admin report: %v", err)
		http.Error(w, "Ошибка сервера при получении отчета", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("Error encoding admin report JSON: %v", err)
	}
}
