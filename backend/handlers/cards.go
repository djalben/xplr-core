package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// GetCardTypesHandler returns available card categories with conditions.
// GET /api/v1/cards/types (public or protected)
func GetCardTypesHandler(w http.ResponseWriter, r *http.Request) {
	types := []map[string]interface{}{
		{
			"category":    "arbitrage",
			"label":       "Для рекламы",
			"description": "Карты для оплаты рекламных кабинетов Facebook, Google, TikTok и трекеров",
			"issue_fee":   "5.00",
			"monthly_fee": "2.00",
			"currency":    "USD",
		},
		{
			"category":    "travel",
			"label":       "Для путешествий",
			"description": "Карты для бронирования отелей, авиабилетов и аренды авто",
			"issue_fee":   "3.00",
			"monthly_fee": "1.50",
			"currency":    "USD",
		},
		{
			"category":    "services",
			"label":       "Для зарубежных сервисов",
			"description": "Карты для подписок, онлайн-сервисов и зарубежных покупок",
			"issue_fee":   "2.00",
			"monthly_fee": "1.00",
			"currency":    "USD",
			"validity":    "1 year",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types)
}

// MockCardDetailsHandler returns simulated card details (PAN, CVV, expiry) for the authorized owner.
// GET /api/v1/user/cards/{id}/mock-details
func MockCardDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

	card, err := repository.GetCardByID(cardID)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}
	if card.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Generate deterministic mock data from card ID and BIN
	fullNumber := card.BIN + "00" + strings.Repeat("0", 6-len(card.Last4Digits)) + card.Last4Digits
	if len(fullNumber) < 16 {
		fullNumber = fullNumber + strings.Repeat("0", 16-len(fullNumber))
	}
	if len(fullNumber) > 16 {
		fullNumber = fullNumber[:16]
	}
	cvv := strconv.Itoa(100 + (cardID*7)%900)
	expiryMonth := (cardID%12 + 1)
	expiryYear := 2027
	expiry := strconv.Itoa(expiryMonth) + "/" + strconv.Itoa(expiryYear)
	if expiryMonth < 10 {
		expiry = "0" + expiry
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"card_id":     card.ID,
		"full_number": fullNumber,
		"cvv":         cvv,
		"expiry":      expiry,
		"card_type":   card.CardType,
		"bin":         card.BIN,
		"last_4":      card.Last4Digits,
	})
}

// UpdateCardSpendLimitHandler handles PATCH /api/v1/user/cards/{id}/limit
func UpdateCardSpendLimitHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	cardID, err := strconv.Atoi(vars["id"])
	if err != nil || cardID <= 0 {
		http.Error(w, "invalid card id", http.StatusBadRequest)
		return
	}
	var req struct {
		Limit decimal.Decimal `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Limit.LessThan(decimal.Zero) {
		http.Error(w, "limit must be >= 0", http.StatusBadRequest)
		return
	}
	if err := repository.UpdateCardSpendLimit(cardID, userID, req.Limit); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"card_id": cardID,
		"limit":   req.Limit.String(),
		"message": "Spend limit updated successfully",
	})
}

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
	case "BLOCKED", "ACTIVE", "FROZEN", "CLOSED":
		status = s
	case "BLOCK":
		status = "BLOCKED"
	case "UNBLOCK", "ACTIVATE":
		status = "ACTIVE"
	case "FREEZE":
		status = "FROZEN"
	case "CLOSE", "DELETE":
		status = "CLOSED"
	default:
		lower := strings.ToLower(req.Status)
		switch lower {
		case "blocked":
			status = "BLOCKED"
		case "active":
			status = "ACTIVE"
		case "frozen", "freeze":
			status = "FROZEN"
		case "closed", "close", "delete":
			status = "CLOSED"
		default:
			http.Error(w, "invalid status: use active, blocked, frozen or closed", http.StatusBadRequest)
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

	if req.Count < 1 {
		req.Count = 1
	}
	if req.Count > 100 {
		http.Error(w, "Максимум 100 карт за раз", http.StatusBadRequest)
		return
	}

	// Calculate fee per card based on category
	feePerCard := decimal.NewFromFloat(5.00) // arbitrage default
	cat := strings.ToLower(req.Category)
	switch cat {
	case "travel":
		feePerCard = decimal.NewFromFloat(3.00)
	case "services":
		feePerCard = decimal.NewFromFloat(2.00)
	}

	// If price_rub is provided (personal cards), convert RUB → USD
	if req.PriceRub.GreaterThan(decimal.Zero) {
		rate, err := repository.GetFinalRate("RUB", "USD")
		if err == nil && rate.GreaterThan(decimal.Zero) {
			feePerCard = req.PriceRub.Div(rate).Round(2)
		}
	}

	totalFee := feePerCard.Mul(decimal.NewFromInt(int64(req.Count)))

	// Deduct balance before issuing
	if totalFee.GreaterThan(decimal.Zero) {
		details := "Card issue fee: " + strconv.Itoa(req.Count) + "x " + cat + " @ $" + feePerCard.StringFixed(2)
		if err := repository.DeductBalance(userID, totalFee, details); err != nil {
			if strings.Contains(err.Error(), "недостаточно средств") {
				http.Error(w, err.Error(), http.StatusPaymentRequired)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
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
		"card_id":   cardID,
		"enabled":   req.Enabled,
		"threshold": req.Threshold.String(),
		"amount":    req.Amount.String(),
		"message":   "Auto-replenishment settings updated successfully",
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
