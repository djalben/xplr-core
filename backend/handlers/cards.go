package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/aalabin/xplr/models"
	"github.com/aalabin/xplr/repository"
	"github.com/aalabin/xplr/middleware"
	"github.com/shopspring/decimal"
)

// patchCardStatusRequest is the JSON body for PATCH /user/cards/:id/status
type patchCardStatusRequest struct {
	Status string `json:"status"`
}

// GetUserCardsHandler handles GET /api/v1/user/cards
func GetUserCardsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cards, err := repository.GetUserCards(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

// PatchCardStatusHandler handles PATCH /api/v1/user/cards/:id/status
func PatchCardStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		http.Error(w, "missing card id", http.StatusBadRequest)
		return
	}
	cardID, err := strconv.Atoi(idStr)
	if err != nil || cardID <= 0 {
		http.Error(w, "invalid card id", http.StatusBadRequest)
		return
	}

	var req patchCardStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	s := strings.TrimSpace(strings.ToUpper(req.Status))
	var status string
	switch s {
	case "BLOCKED", "ACTIVE":
		status = s
	case "BLOCK":
		status = "BLOCKED"
	case "UNBLOCK", "ACTIVATE":
		status = "ACTIVE"
	default:
		// support lowercase "blocked" / "active"
		lower := strings.ToLower(req.Status)
		if lower == "blocked" {
			status = "BLOCKED"
		} else if lower == "active" {
			status = "ACTIVE"
		} else {
			http.Error(w, "invalid status: use blocked or active", http.StatusBadRequest)
			return
		}
	}

	if err := repository.UpdateCardStatus(cardID, userID, status); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"card_id": cardID,
		"status":  status,
	})
}

func MassIssueCardsHandler(w http.ResponseWriter, r *http.Request) {
	// Используем константу из middleware для получения ID пользователя
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	var req models.MassIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат", http.StatusBadRequest)
		return
	}

	response, err := repository.IssueCards(userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// SetCardAutoReplenishmentHandler - POST /api/v1/user/cards/{id}/auto-replenishment
func SetCardAutoReplenishmentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		http.Error(w, "missing card id", http.StatusBadRequest)
		return
	}
	cardID, err := strconv.Atoi(idStr)
	if err != nil || cardID <= 0 {
		http.Error(w, "invalid card id", http.StatusBadRequest)
		return
	}

	var req models.AutoReplenishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Enabled {
		if req.Threshold.LessThanOrEqual(decimal.Zero) {
			http.Error(w, "threshold must be greater than 0", http.StatusBadRequest)
			return
		}
		if req.Amount.LessThanOrEqual(decimal.Zero) {
			http.Error(w, "amount must be greater than 0", http.StatusBadRequest)
			return
		}
	}

	if err := repository.UpdateCardAutoReplenishment(cardID, userID, req.Enabled, req.Threshold, req.Amount); err != nil {
		if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"card_id": cardID,
		"enabled": req.Enabled,
		"threshold": req.Threshold.String(),
		"amount": req.Amount.String(),
		"message": "Auto-replenishment settings updated successfully",
	})
}

// UnsetCardAutoReplenishmentHandler - DELETE /api/v1/user/cards/{id}/auto-replenishment
func UnsetCardAutoReplenishmentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		http.Error(w, "missing card id", http.StatusBadRequest)
		return
	}
	cardID, err := strconv.Atoi(idStr)
	if err != nil || cardID <= 0 {
		http.Error(w, "invalid card id", http.StatusBadRequest)
		return
	}

	// Отключаем автопополнение (устанавливаем enabled = false)
	if err := repository.UpdateCardAutoReplenishment(cardID, userID, false, decimal.Zero, decimal.Zero); err != nil {
		if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"card_id": cardID,
		"message": "Auto-replenishment disabled successfully",
	})
}