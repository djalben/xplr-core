package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
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

	// Generate deterministic mock data from card provider_card_id (which is the full 16-digit PAN)
	fullNumber := card.ProviderCardID
	if len(fullNumber) < 16 {
		// Fallback: reconstruct from BIN + padding + last4
		fullNumber = card.BIN + strings.Repeat("0", 16-len(card.BIN)-len(card.Last4Digits)) + card.Last4Digits
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

	// Fetch card details before status change for notification (last4 + balance for refund)
	var cardLast4 string
	var cardBalanceBefore decimal.Decimal
	if card, err := repository.GetCardByID(cardID); err == nil {
		cardLast4 = card.Last4Digits
		cardBalanceBefore = card.CardBalance
	}

	if err := repository.UpdateCardStatus(cardID, userID, status); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[EVENT] User %d performed card_status_change (card=%d, last4=%s, status=%s). Triggering notifications...", userID, cardID, cardLast4, status)

	// Notify user about card status change
	statusLabels := map[string]string{
		"FROZEN":  "❄️ Карта заморожена",
		"ACTIVE":  "✅ Карта активирована",
		"BLOCKED": "🔒 Карта заблокирована",
		"CLOSED":  "❌ Карта закрыта",
	}
	label := statusLabels[status]
	if label == "" {
		label = status
	}

	var msg string
	switch status {
	case "CLOSED":
		msg = fmt.Sprintf("❌ <b>Карта закрыта</b>\n\nВаша карта *%s успешно закрыта и аннулирована.\n\n<a href=\"https://xplr.pro/cards\">Открыть карты</a>", cardLast4)
	case "FROZEN":
		msg = fmt.Sprintf("❄️ <b>Карта заморожена</b>\n\nСтатус вашей карты *%s изменён на: <b>FROZEN</b>\n\n<a href=\"https://xplr.pro/cards\">Открыть карты</a>", cardLast4)
	case "ACTIVE":
		msg = fmt.Sprintf("✅ <b>Карта разморожена</b>\n\nСтатус вашей карты *%s изменён на: <b>ACTIVE</b>\n\n<a href=\"https://xplr.pro/cards\">Открыть карты</a>", cardLast4)
	case "BLOCKED":
		msg = fmt.Sprintf("🔒 <b>Карта заблокирована</b>\n\nСтатус вашей карты *%s изменён на: <b>BLOCKED</b>\n\n<a href=\"https://xplr.pro/cards\">Открыть карты</a>", cardLast4)
	default:
		msg = fmt.Sprintf("%s\n\nКарта *%s\nСтатус: <b>%s</b>\n\n<a href=\"https://xplr.pro/cards\">Открыть карты</a>", label, cardLast4, status)
	}
	go service.NotifyUser(userID, label, msg)

	// Refund notification: if card had balance and was closed/blocked, the balance was returned to wallet
	if (status == "CLOSED" || status == "BLOCKED") && cardBalanceBefore.GreaterThan(decimal.Zero) {
		log.Printf("[EVENT] User %d performed card_refund_to_wallet (card=%d, last4=%s, refunded=$%s). Triggering notifications...",
			userID, cardID, cardLast4, cardBalanceBefore.StringFixed(2))
		go service.NotifyUser(userID, "Возврат средств с карты",
			fmt.Sprintf("💰 <b>Возврат средств</b>\n\n"+
				"Списание с карты *%s произведено.\n"+
				"Средства <b>$%s</b> возвращены на основной баланс.\n\n"+
				"<a href=\"https://xplr.pro/wallet\">Открыть кошелёк</a>",
				cardLast4, cardBalanceBefore.StringFixed(2)))
	}

	// Return updated wallet balance so frontend can refresh instantly
	resp := map[string]interface{}{
		"card_id": cardID,
		"status":  status,
	}
	if ib, err := repository.GetInternalBalance(userID); err == nil {
		resp["wallet_balance"] = ib.MasterBalance
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
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

	// Check tier-based card limit
	var tier string
	var tierExpiresAt sql.NullTime
	var currentCardCount int
	err := GlobalDB.QueryRow(`
		SELECT COALESCE(tier, 'standard'), tier_expires_at, 
		       (SELECT COUNT(*) FROM cards WHERE user_id = $1)
		FROM users WHERE id = $1
	`, userID).Scan(&tier, &tierExpiresAt, &currentCardCount)
	if err != nil {
		http.Error(w, "Ошибка проверки лимита", http.StatusInternalServerError)
		return
	}

	// Determine card limit based on tier
	cardLimit := 3 // standard tier
	if tier == "gold" && tierExpiresAt.Valid && tierExpiresAt.Time.After(time.Now()) {
		cardLimit = 15 // gold tier
	}

	if currentCardCount+req.Count > cardLimit {
		http.Error(w, fmt.Sprintf("Превышен лимит карт. Ваш лимит: %d карт (текущих: %d). Обновитесь до Gold tier для лимита 15 карт.", cardLimit, currentCardCount), http.StatusForbidden)
		return
	}

	// Calculate fee in USD (master_balance is now in USD)
	cat := strings.ToLower(req.Category)
	if cat == "" {
		cat = "arbitrage"
	}

	var feeUSD decimal.Decimal
	if req.PriceUSD.GreaterThan(decimal.Zero) {
		// Personal cards: fixed USD price sent by frontend
		feeUSD = req.PriceUSD
	} else {
		// Arbitrage cards: default USD fees
		feeUSD = decimal.NewFromFloat(5.00)
		switch cat {
		case "travel":
			feeUSD = decimal.NewFromFloat(3.00)
		case "services":
			feeUSD = decimal.NewFromFloat(2.00)
		}
	}
	totalFeeUSD := feeUSD.Mul(decimal.NewFromInt(int64(req.Count)))

	// Deduct from wallet (internal_balances.master_balance, USD)
	if totalFeeUSD.GreaterThan(decimal.Zero) {
		details := "Card issue fee: " + strconv.Itoa(req.Count) + "x " + cat + " — $" + totalFeeUSD.StringFixed(2)
		if err := repository.DeductWalletBalance(userID, totalFeeUSD, details); err != nil {
			if strings.Contains(err.Error(), "недостаточно средств") || strings.Contains(err.Error(), "кошелёк не найден") {
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

	log.Printf("[EVENT] User %d performed card_issue (count=%d, category=%s, fee=$%s). Triggering notifications...", userID, req.Count, cat, totalFeeUSD.StringFixed(2))

	// Notify user about card issue (single unified notification — TG + Email)
	go service.NotifyUser(userID, "Карта выпущена",
		fmt.Sprintf("💳 <b>Карта успешно выпущена!</b>\n\n"+
			"📦 <b>Количество:</b> %d\n"+
			"🏷 <b>Категория:</b> %s\n"+
			"💰 <b>Комиссия:</b> $%s\n"+
			"� <b>Дневной лимит:</b> $%s\n\n"+
			"Карта уже доступна в <a href=\"https://xplr.pro/cards\">личном кабинете</a>.",
			req.Count, cat, totalFeeUSD.StringFixed(2), req.DailyLimit.StringFixed(2)))

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
