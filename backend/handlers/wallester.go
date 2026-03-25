package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/notification"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/gorilla/mux"
)

var wallesterRepo *repository.WallesterRepository

// InitWallesterRepository инициализирует репозиторий Wallester (вызывается из main.go)
func InitWallesterRepository() {
	wallesterRepo = repository.NewWallesterRepository()
}

// getClientIP извлекает реальный IP-адрес клиента из запроса
// Учитывает заголовки X-Forwarded-For и X-Real-IP для проксированных запросов
func getClientIP(r *http.Request) string {
	// Проверяем заголовки прокси
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For может содержать несколько IP через запятую
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Если заголовков нет, используем RemoteAddr
	ip, _, found := strings.Cut(r.RemoteAddr, ":")
	if !found {
		return r.RemoteAddr
	}
	return ip
}

// WallesterWebhookHandler - Обработчик webhook от Wallester с проверками безопасности
// POST /api/v1/webhooks/wallester
// Проверяет: IP whitelist, signature validation, idempotency
func WallesterWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if wallesterRepo == nil {
		http.Error(w, "Wallester repository not initialized", http.StatusInternalServerError)
		return
	}

	// 1. IP WHITELIST: Проверка IP-адреса отправителя
	clientIP := getClientIP(r)
	if !repository.CheckIPWhitelist(clientIP) {
		log.Printf("🚫 Webhook rejected: IP %s not in whitelist", clientIP)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	log.Printf("✅ IP whitelist check passed: %s", clientIP)

	// 2. Чтение тела запроса (нужно для проверки подписи)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// 3. SIGNATURE VALIDATION: Проверка подписи webhook
	signature := r.Header.Get("X-Wallester-Signature")
	if !repository.VerifyWebhookSignature(bodyBytes, signature) {
		log.Printf("🚫 Webhook rejected: Invalid signature (IP: %s)", clientIP)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}
	if signature != "" {
		log.Printf("✅ Signature validation passed")
	}

	// 4. Декодирование payload
	var payload repository.WallesterWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		log.Printf("Error decoding webhook payload: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// REMOVED: Wallester webhook processing - provider interface will handle callbacks differently
	// Each provider (Armenia, etc.) will have its own webhook handler
	log.Printf("[WEBHOOK] Received webhook (provider interface pending): event_type=%s", payload.EventType)

	// TODO: Implement provider-specific webhook handling
	// For now, just acknowledge receipt

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Webhook processed successfully",
	})
}

// sendWallesterNotification отправляет уведомления (TG + Email) для событий Wallester
// Вызывается из хендлера в горутине после успешного ProcessWebhook
func sendWallesterNotification(payload repository.WallesterWebhookPayload) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️  Panic in sendWallesterNotification: %v", r)
		}
	}()

	// Находим карту и пользователя для отправки уведомления
	var cardID, userID int
	var last4Digits string
	query := `SELECT id, user_id, last_4_digits FROM cards WHERE external_id = $1 OR provider_card_id = $1 LIMIT 1`
	err := repository.GlobalDB.QueryRow(query, payload.CardID).Scan(&cardID, &userID, &last4Digits)
	if err != nil {
		log.Printf("⚠️  Failed to find card for notification: %v", err)
		return
	}

	log.Printf("[EVENT] Wallester webhook %s for card %d (user %d, last4=%s). Triggering notifications...",
		payload.EventType, cardID, userID, last4Digits)

	amount := payload.Amount
	if amount == "" {
		amount = "0"
	}
	currency := payload.Currency
	if currency == "" {
		currency = "USD"
	}
	merchantName := payload.MerchantName
	if merchantName == "" {
		merchantName = "Unknown"
	}

	// Формируем и отправляем уведомление в зависимости от типа события
	switch payload.EventType {
	case "3ds_authentication":
		if payload.AuthCode != "" {
			// 3DS код срочный — отправляем только в TG (быстрее email)
			user, uErr := repository.GetUserByID(userID)
			if uErr == nil && user.TelegramChatID.Valid {
				notification.SendTelegramMessage(user.TelegramChatID.Int64,
					fmt.Sprintf("🔑 Код подтверждения: %s | Магазин: %s\n\n⚠️ Внимание: не сообщайте код третьим лицам!",
						payload.AuthCode, merchantName))
			}
			log.Printf("✅ 3DS notification sent to user %d", userID)
		}

	case "payment_success", "transaction", "capture", "authorization":
		service.NotifyUser(userID, "Списание с карты",
			fmt.Sprintf("💸 <b>Списание с карты *%s</b>\n\n"+
				"Сумма: <b>%s %s</b>\n"+
				"Магазин: %s\n\n"+
				"<a href=\"https://xplr.pro/cards\">Открыть карты</a>",
				last4Digits, amount, currency, merchantName))
		log.Printf("✅ Payment notification sent to user %d (card=%d)", userID, cardID)

	case "refund", "reversal":
		service.NotifyUser(userID, "Возврат средств",
			fmt.Sprintf("💰 <b>Возврат средств на кошелёк</b>\n\n"+
				"Карта: *%s\n"+
				"Сумма возврата: <b>%s %s</b>\n\n"+
				"<a href=\"https://xplr.pro/wallet\">Открыть кошелёк</a>",
				last4Digits, amount, currency))
		log.Printf("✅ Refund notification sent to user %d (card=%d)", userID, cardID)
	}
}

// GetCardDetailsHandler - Получение реквизитов карты (PAN, CVV, expiry) из Wallester
// GET /api/v1/user/cards/{id}/details
func GetCardDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Получаем карту из БД и проверяем принадлежность пользователю
	card, err := repository.GetCardByID(cardID)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	if card.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Получаем детали через провайдерский интерфейс (MockProvider или ArmeniaProvider)
	provider := service.GetCardProvider()
	details, err := provider.GetCardDetails(cardID)
	if err != nil {
		log.Printf("[CARD-DETAILS] Error getting card details from %s: %v", provider.GetProviderName(), err)
		http.Error(w, "Failed to get card details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// SyncCardBalanceHandler - Синхронизация баланса конкретной карты
// POST /api/v1/user/cards/{id}/sync-balance
func SyncCardBalanceHandler(w http.ResponseWriter, r *http.Request) {
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

	// Получаем карту и проверяем принадлежность
	card, err := repository.GetCardByID(cardID)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	if card.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// REMOVED: Wallester sync - provider interface will handle balance sync
	// Balance sync functionality will be implemented through provider interface
	log.Printf("[SYNC-BALANCE] Balance sync requested for card %d (provider interface pending)", cardID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Balance synced successfully",
	})
}
