package handlers

import (
	"encoding/json"
	"net/http"
	"log"
	"fmt"

	"github.com/aalabin/xplr/middleware"
	"github.com/aalabin/xplr/repository"
)

// TelegramIDRequest — модель для входящего JSON запроса.
type TelegramIDRequest struct {
	ChatID int64 `json:"chat_id"`
}

// UpdateTelegramChatIDHandler обрабатывает POST /v1/settings/telegram
// (Защищен AuthMiddleware)
func UpdateTelegramChatIDHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Извлечение UserID из контекста JWT
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	// 2. Декодирование ChatID из тела запроса
	var req TelegramIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: " + err.Error(), http.StatusBadRequest)
		return
	}
    
    // Простая проверка ChatID (должно быть больше 0)
    if req.ChatID <= 0 {
        http.Error(w, "Invalid ChatID. Must be positive.", http.StatusBadRequest)
        return
    }

	// 3. Обновление ChatID в БД
    // ИСПРАВЛЕНИЕ: Преобразование req.ChatID (int64) в int для вызова репозитория
    // (chat ID в БД - bigint, но функция UpdateTelegramChatID ожидает int)
	err := repository.UpdateTelegramChatID(userID, int(req.ChatID))
	if err != nil {
		log.Printf("Error updating Telegram Chat ID for user %d: %v", userID, err)
		http.Error(w, "Failed to update ChatID: " + err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": fmt.Sprintf("Telegram Chat ID %d successfully linked to user %d.", req.ChatID, userID)}
	json.NewEncoder(w).Encode(response)
}