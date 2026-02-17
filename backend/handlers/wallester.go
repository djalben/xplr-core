package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/aalabin/xplr/backend/middleware"
	"github.com/aalabin/xplr/backend/notification"
	"github.com/aalabin/xplr/backend/repository"
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

	// 5. Обработка webhook (включает проверку idempotency внутри ProcessWebhook)
	if err := wallesterRepo.ProcessWebhook(payload); err != nil {
		log.Printf("Error processing webhook: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Отправка Telegram-уведомлений для событий 3DS и успешных платежей
	// (дополнительно к уведомлениям в ProcessWebhook для явности и контроля)
	if payload.EventType == "3ds_authentication" || payload.EventType == "payment_success" {
		sendWallesterNotification(payload)
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Webhook processed successfully",
	})
}

// sendWallesterNotification отправляет Telegram-уведомления для событий Wallester
// Вызывается из хендлера для явного контроля уведомлений
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

	// Получаем пользователя
	user, err := repository.GetUserByID(userID)
	if err != nil {
		log.Printf("⚠️  Failed to get user for notification: %v", err)
		return
	}

	if !user.TelegramChatID.Valid {
		log.Printf("⚠️  User %d has no Telegram chat ID, skipping notification", userID)
		return
	}

	// Формируем и отправляем уведомление в зависимости от типа события
	switch payload.EventType {
	case "3ds_authentication":
		if payload.AuthCode != "" {
			merchantName := payload.MerchantName
			if merchantName == "" {
				merchantName = "Unknown"
			}
			message := fmt.Sprintf(
				"🔑 Код подтверждения: %s | Магазин: %s\n\n⚠️ Внимание: не сообщайте код третьим лицам!",
				payload.AuthCode,
				merchantName,
			)
			notification.SendTelegramMessage(user.TelegramChatID.Int64, message)
			log.Printf("✅ 3DS notification sent to user %d", userID)
		}

	case "payment_success":
		amount := payload.Amount
		if amount == "" {
			amount = "0"
		}
		currency := payload.Currency
		if currency == "" {
			currency = "RUB"
		}
		merchantName := payload.MerchantName
		if merchantName == "" {
			merchantName = "Unknown"
		}

		// Получаем текущий баланс пользователя
		var balanceRub string
		err := repository.GlobalDB.QueryRow("SELECT balance_rub FROM users WHERE id = $1", userID).Scan(&balanceRub)
		if err != nil {
			log.Printf("⚠️  Failed to get balance for notification: %v", err)
			balanceRub = "N/A"
		}

		message := fmt.Sprintf(
			"💸 Списание: %s %s | Карта: *%s | Магазин: %s\n\nВаш новый баланс: %s₽",
			amount,
			currency,
			last4Digits,
			merchantName,
			balanceRub,
		)
		notification.SendTelegramMessage(user.TelegramChatID.Int64, message)
		log.Printf("✅ Payment success notification sent to user %d", userID)
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

	if wallesterRepo == nil {
		http.Error(w, "Wallester repository not initialized", http.StatusInternalServerError)
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

	// Используем external_id или provider_card_id для запроса к Wallester
	externalID := card.ExternalID
	if externalID == "" {
		externalID = card.ProviderCardID
	}

	if externalID == "" {
		http.Error(w, "Card external ID not found", http.StatusBadRequest)
		return
	}

	// Получаем детали из Wallester
	details, err := wallesterRepo.GetCardDetails(externalID)
	if err != nil {
		log.Printf("Error getting card details from Wallester: %v", err)
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

	if wallesterRepo == nil {
		http.Error(w, "Wallester repository not initialized", http.StatusInternalServerError)
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

	externalID := card.ExternalID
	if externalID == "" {
		externalID = card.ProviderCardID
	}

	if externalID == "" {
		http.Error(w, "Card external ID not found", http.StatusBadRequest)
		return
	}

	// Синхронизация баланса
	if err := wallesterRepo.SyncBalance(cardID, externalID); err != nil {
		log.Printf("Error syncing balance: %v", err)
		http.Error(w, "Failed to sync balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"message": "Balance synced successfully",
	})
}
