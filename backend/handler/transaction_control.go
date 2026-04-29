package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/notification"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// ══════════════════════════════════════════════════════════════
// 1. Smart Subscription Filter (Anti-Drain)
// ══════════════════════════════════════════════════════════════

// PATCH /api/v1/user/cards/{id}/auto-pay — toggle is_auto_pay_enabled
func ToggleAutoPayHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	cardID, _ := strconv.Atoi(vars["id"])
	if cardID <= 0 {
		http.Error(w, "invalid card id", http.StatusBadRequest)
		return
	}

	card, err := repository.GetCardByID(cardID)
	if err != nil || card.UserID != userID {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = GlobalDB.Exec(`UPDATE cards SET is_auto_pay_enabled = $1 WHERE id = $2`, req.Enabled, cardID)
	if err != nil {
		log.Printf("[AUTO-PAY] ❌ Failed to update card %d: %v", cardID, err)
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	status := "disabled"
	if req.Enabled {
		status = "enabled"
	}
	log.Printf("[AUTO-PAY] Card %d auto-pay %s (user %d)", cardID, status, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":              "ok",
		"is_auto_pay_enabled": req.Enabled,
	})
}

// CheckRecurringAllowed checks if a recurring/subscription transaction should be declined.
// Returns true if allowed, false if should be declined.
func CheckRecurringAllowed(cardID int, merchantName string, isRecurring bool) bool {
	if !isRecurring {
		return true // Not a subscription — always allow
	}

	// 1. Check global auto-pay toggle on card
	var autoPayEnabled bool
	err := GlobalDB.QueryRow(
		`SELECT COALESCE(is_auto_pay_enabled, TRUE) FROM cards WHERE id = $1`, cardID,
	).Scan(&autoPayEnabled)
	if err != nil {
		log.Printf("[RECURRING-CHECK] ❌ DB error for card %d: %v", cardID, err)
		return true // On error, allow (fail-open)
	}

	if !autoPayEnabled {
		log.Printf("[RECURRING-CHECK] 🚫 Card %d has auto-pay DISABLED — declining recurring tx from %q", cardID, merchantName)
		return false
	}

	// 2. Check per-merchant block
	if merchantName != "" {
		var isBlocked bool
		err := GlobalDB.QueryRow(
			`SELECT COALESCE(is_blocked, FALSE) FROM merchant_blocks WHERE card_id = $1 AND LOWER(merchant_name) = LOWER($2)`,
			cardID, merchantName,
		).Scan(&isBlocked)
		if err == nil && isBlocked {
			log.Printf("[RECURRING-CHECK] 🚫 Merchant %q blocked on card %d", merchantName, cardID)
			return false
		}
	}

	return true
}

// ══════════════════════════════════════════════════════════════
// 2. Subscription Dashboard
// ══════════════════════════════════════════════════════════════

// TrackMerchantSubscription — called after successful charge to auto-track merchants
func TrackMerchantSubscription(cardID, userID int, merchantName, amount, currency string) {
	if merchantName == "" || GlobalDB == nil {
		return
	}
	_, err := GlobalDB.Exec(`
		INSERT INTO card_subscriptions (card_id, user_id, merchant_name, last_amount, last_currency, charge_count, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, 1, NOW())
		ON CONFLICT (card_id, merchant_name) DO UPDATE SET
			last_amount = EXCLUDED.last_amount,
			last_currency = EXCLUDED.last_currency,
			charge_count = card_subscriptions.charge_count + 1,
			last_seen_at = NOW()
	`, cardID, userID, merchantName, amount, currency)
	if err != nil {
		log.Printf("[SUBSCRIPTION-TRACK] ❌ Failed to track merchant %q for card %d: %v", merchantName, cardID, err)
	}
}

// GET /api/v1/user/cards/{id}/subscriptions — list tracked subscriptions for a card
func CardSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	cardID, _ := strconv.Atoi(vars["id"])
	if cardID <= 0 {
		http.Error(w, "invalid card id", http.StatusBadRequest)
		return
	}

	card, err := repository.GetCardByID(cardID)
	if err != nil || card.UserID != userID {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	rows, err := GlobalDB.Query(`
		SELECT id, merchant_name, COALESCE(last_amount, 0), COALESCE(last_currency, 'USD'),
			charge_count, is_allowed, first_seen_at, last_seen_at
		FROM card_subscriptions
		WHERE card_id = $1
		ORDER BY last_seen_at DESC
	`, cardID)
	if err != nil {
		http.Error(w, "Failed to fetch subscriptions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Sub struct {
		ID           int       `json:"id"`
		MerchantName string    `json:"merchant_name"`
		LastAmount   string    `json:"last_amount"`
		LastCurrency string    `json:"last_currency"`
		ChargeCount  int       `json:"charge_count"`
		IsAllowed    bool      `json:"is_allowed"`
		FirstSeenAt  time.Time `json:"first_seen_at"`
		LastSeenAt   time.Time `json:"last_seen_at"`
	}

	var subs []Sub
	for rows.Next() {
		var s Sub
		if err := rows.Scan(&s.ID, &s.MerchantName, &s.LastAmount, &s.LastCurrency,
			&s.ChargeCount, &s.IsAllowed, &s.FirstSeenAt, &s.LastSeenAt); err != nil {
			continue
		}
		subs = append(subs, s)
	}
	if subs == nil {
		subs = []Sub{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscriptions":       subs,
		"is_auto_pay_enabled": card.IsAutoPayEnabled,
	})
}

// PATCH /api/v1/user/cards/{id}/subscriptions/{subId} — toggle subscription allow/block
func ToggleSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	cardID, _ := strconv.Atoi(vars["id"])
	subID, _ := strconv.Atoi(vars["subId"])
	if cardID <= 0 || subID <= 0 {
		http.Error(w, "invalid ids", http.StatusBadRequest)
		return
	}

	card, err := repository.GetCardByID(cardID)
	if err != nil || card.UserID != userID {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	var req struct {
		IsAllowed bool `json:"is_allowed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update subscription allowed status
	_, err = GlobalDB.Exec(
		`UPDATE card_subscriptions SET is_allowed = $1 WHERE id = $2 AND card_id = $3`,
		req.IsAllowed, subID, cardID,
	)
	if err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	// Sync to merchant_blocks table
	var merchantName string
	GlobalDB.QueryRow(`SELECT merchant_name FROM card_subscriptions WHERE id = $1`, subID).Scan(&merchantName)
	if merchantName != "" {
		if req.IsAllowed {
			GlobalDB.Exec(`DELETE FROM merchant_blocks WHERE card_id = $1 AND LOWER(merchant_name) = LOWER($2)`, cardID, merchantName)
		} else {
			GlobalDB.Exec(`
				INSERT INTO merchant_blocks (card_id, merchant_name, is_blocked)
				VALUES ($1, $2, TRUE)
				ON CONFLICT (card_id, merchant_name) DO UPDATE SET is_blocked = TRUE
			`, cardID, merchantName)
		}
	}

	log.Printf("[SUBSCRIPTION] Card %d merchant %q → allowed=%v (user %d)", cardID, merchantName, req.IsAllowed, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ══════════════════════════════════════════════════════════════
// 3. 3DS/SMS Notification Hub
// ══════════════════════════════════════════════════════════════

// Armenian bank SMS patterns (OTP code extraction)
var smsCodePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:код|code|OTP|pin)[:\s]*(\d{4,6})`),
	regexp.MustCompile(`(?i)(\d{4,6})\s*(?:—|–|-|is your|ваш|kod)`),
	regexp.MustCompile(`(?i)(?:подтвержд|verif|confirm|authen)[^\d]*(\d{4,6})`),
	regexp.MustCompile(`(?i)3[dD][sS][^\d]*(\d{4,6})`),
	regexp.MustCompile(`(?i)(?:password|пароль|parol)[:\s]*(\d{4,6})`),
	// Armenian-specific: Ameriabank, Ardshinbank, ACBA, IDBank patterns
	regexp.MustCompile(`(?i)(?:Ameria|Ardshin|ACBA|IDBank|Evoca|Inecobank)[^\d]*(\d{4,6})`),
	// Generic: standalone 6-digit code in SMS
	regexp.MustCompile(`\b(\d{6})\b`),
}

// ExtractSMSCode extracts a 4-6 digit OTP code from an SMS message
func ExtractSMSCode(message string) string {
	for _, pattern := range smsCodePatterns {
		matches := pattern.FindStringSubmatch(message)
		if len(matches) >= 2 {
			code := matches[1]
			if len(code) >= 4 && len(code) <= 6 {
				return code
			}
		}
	}
	return ""
}

// POST /api/v1/webhooks/sms-receiver — receive SMS/3DS codes from Armenian bank
func SMSReceiverWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      int    `json:"user_id"`
		CardID      int    `json:"card_id"`
		Message     string `json:"message"`
		Sender      string `json:"sender"`
		PhoneNumber string `json:"phone_number"`
		Timestamp   string `json:"timestamp"`
		// Alternative: direct code field (if pre-parsed by gateway)
		Code         string `json:"code"`
		MerchantName string `json:"merchant_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID <= 0 {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	// Extract code from message (or use pre-parsed code)
	code := req.Code
	if code == "" && req.Message != "" {
		code = ExtractSMSCode(req.Message)
	}
	if code == "" {
		log.Printf("[SMS-HUB] ⚠️ No code extracted from message: %q (user %d)", req.Message, req.UserID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "no_code_found"})
		return
	}

	merchant := req.MerchantName
	if merchant == "" && req.Message != "" {
		// Try to extract merchant from message
		merchant = extractMerchantFromSMS(req.Message)
	}

	log.Printf("[SMS-HUB] ✅ Code %s received for user %d (card=%d, merchant=%q)", code, req.UserID, req.CardID, merchant)

	// 1. Save to DB
	var smsID int
	err := GlobalDB.QueryRow(`
		INSERT INTO sms_codes (user_id, card_id, code, merchant_name, raw_message)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, req.UserID, req.CardID, code, merchant, req.Message).Scan(&smsID)
	if err != nil {
		log.Printf("[SMS-HUB] ❌ DB insert error: %v", err)
	}

	// 2. Deliver via WebSocket (zero-latency)
	deliveredWS := ThreeDSHub.Deliver(req.UserID, SMSCodeMessage{
		ID:           smsID,
		Code:         code,
		MerchantName: merchant,
		CardID:       req.CardID,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	})
	if deliveredWS {
		GlobalDB.Exec(`UPDATE sms_codes SET delivered_ws = TRUE WHERE id = $1`, smsID)
	}

	// 3. Deliver via Telegram (if linked)
	deliveredTG := false
	user, uErr := repository.GetUserByID(req.UserID)
	if uErr == nil && user.TelegramChatID.Valid {
		tgMsg := fmt.Sprintf("🔑 <b>3DS Код: %s</b>", code)
		if merchant != "" {
			tgMsg += fmt.Sprintf("\nМагазин: %s", merchant)
		}
		tgMsg += "\n\n⚠️ Не сообщайте код третьим лицам!"
		notification.SendTelegramMessage(user.TelegramChatID.Int64, tgMsg)
		deliveredTG = true
		GlobalDB.Exec(`UPDATE sms_codes SET delivered_tg = TRUE WHERE id = $1`, smsID)
	}

	log.Printf("[SMS-HUB] Delivery: ws=%v tg=%v (sms_id=%d)", deliveredWS, deliveredTG, smsID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"code":         code,
		"delivered_ws": deliveredWS,
		"delivered_tg": deliveredTG,
	})
}

// extractMerchantFromSMS tries to extract merchant name from SMS text
func extractMerchantFromSMS(message string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:оплата|payment|purchase|покупка)\s+(?:в|at|in)\s+(\S+(?:\s+\S+)?)`),
		regexp.MustCompile(`(?i)(?:магазин|merchant|shop|store)\s*[:\s]+(\S+(?:\s+\S+)?)`),
	}
	for _, p := range patterns {
		m := p.FindStringSubmatch(message)
		if len(m) >= 2 {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

// ══════════════════════════════════════════════════════════════
// WebSocket Hub for real-time 3DS code delivery
// ══════════════════════════════════════════════════════════════

type SMSCodeMessage struct {
	ID           int    `json:"id"`
	Code         string `json:"code"`
	MerchantName string `json:"merchant_name"`
	CardID       int    `json:"card_id"`
	Timestamp    string `json:"timestamp"`
}

type threeDSHub struct {
	mu      sync.RWMutex
	clients map[int][]*websocket.Conn // userID → active WS connections
}

var ThreeDSHub = &threeDSHub{
	clients: make(map[int][]*websocket.Conn),
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Register adds a WebSocket connection for a user
func (h *threeDSHub) Register(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[userID] = append(h.clients[userID], conn)
	h.mu.Unlock()
	log.Printf("[3DS-WS] ✅ Client connected: user %d (total=%d)", userID, len(h.clients[userID]))
}

// Unregister removes a WebSocket connection
func (h *threeDSHub) Unregister(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.clients[userID]
	for i, c := range conns {
		if c == conn {
			h.clients[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
}

// Deliver sends a 3DS code to all WS connections of a user
func (h *threeDSHub) Deliver(userID int, msg SMSCodeMessage) bool {
	h.mu.RLock()
	conns := h.clients[userID]
	h.mu.RUnlock()

	if len(conns) == 0 {
		return false
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"type": "3ds_code",
		"data": msg,
	})

	delivered := false
	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("[3DS-WS] ⚠️ Write error: %v", err)
			conn.Close()
		} else {
			delivered = true
		}
	}
	return delivered
}

// GET /api/v1/user/3ds-ws — WebSocket endpoint for real-time 3DS codes
func ThreeDSWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[3DS-WS] ❌ Upgrade error: %v", err)
		return
	}

	ThreeDSHub.Register(userID, conn)

	// Send pending codes (last 5 minutes, not yet delivered via WS)
	go sendPendingCodes(userID, conn)

	// Keep connection alive; read pump (handles pong/close)
	go func() {
		defer func() {
			ThreeDSHub.Unregister(userID, conn)
			conn.Close()
			log.Printf("[3DS-WS] Client disconnected: user %d", userID)
		}()
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	// Ping ticker to keep connection alive
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()
}

func sendPendingCodes(userID int, conn *websocket.Conn) {
	if GlobalDB == nil {
		return
	}
	rows, err := GlobalDB.Query(`
		SELECT id, code, COALESCE(merchant_name, ''), COALESCE(card_id, 0), created_at
		FROM sms_codes
		WHERE user_id = $1 AND delivered_ws = FALSE AND created_at > NOW() - INTERVAL '5 minutes'
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var msg SMSCodeMessage
		var createdAt time.Time
		if err := rows.Scan(&msg.ID, &msg.Code, &msg.MerchantName, &msg.CardID, &createdAt); err != nil {
			continue
		}
		msg.Timestamp = createdAt.UTC().Format(time.RFC3339)
		payload, _ := json.Marshal(map[string]interface{}{
			"type": "3ds_code",
			"data": msg,
		})
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			break
		}
		GlobalDB.Exec(`UPDATE sms_codes SET delivered_ws = TRUE WHERE id = $1`, msg.ID)
	}
}

// ══════════════════════════════════════════════════════════════
// 4. Telegram Link Check API (for frontend modal)
// ══════════════════════════════════════════════════════════════

// GET /api/v1/user/telegram-status — check if Telegram is linked
func TelegramStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := repository.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	isLinked := user.TelegramChatID.Valid && user.TelegramChatID.Int64 != 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"telegram_linked": isLinked,
	})
}
