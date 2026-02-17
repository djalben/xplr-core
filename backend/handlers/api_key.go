package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aalabin/xplr/backend/repository"
	"github.com/aalabin/xplr/backend/middleware" // <--- ДОБАВЛЕНО: для доступа к UserIDKey
	// УДАЛЕНО: "database/sql" (Не нужен, так как GlobalDB здесь не объявляется)
	// УДАЛЕНО: "github.com/google/uuid" (Не используется)
)

// GlobalDB не объявляется здесь, чтобы избежать ошибки redeclared.
// Предполагается, что var GlobalDB *sql.DB объявлена в handlers/v1.go или другом файле пакета.


// CreateAPIKeyHandler - Хендлер для генерации и сохранения нового API Key
func CreateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Извлечение UserID из контекста JWT
	// ИСПРАВЛЕНО: Используем константу middleware.UserIDKey
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid User ID in context", http.StatusUnauthorized)
		return
	}
	
	// 2. Генерация и сохранение нового ключа
	apiKey, err := repository.GenerateAPIKey(userID) 
	if err != nil {
		log.Printf("Error generating and saving API Key for user %d: %v", userID, err)
		http.Error(w, "Failed to save API Key", http.StatusInternalServerError)
		return
	}

	// 3. Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	
	// Возвращаем только сгенерированный ключ
	json.NewEncoder(w).Encode(map[string]string{"api_key": apiKey})
}