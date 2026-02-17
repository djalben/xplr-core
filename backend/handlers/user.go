package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/shopspring/decimal"
)

// GetMeHandler - Возвращает данные о текущем пользователе (Задача 3.1)
func GetMeHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Извлечение UserID из контекста JWT/API Key
	// Используем UserIDKey, который устанавливает AuthMiddleware
	userID, ok := r.Context().Value(middleware.UserIDKey).(int) 
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	// 2. Получение данных пользователя из БД
	user, err := repository.GetUserByID(userID)
	if err != nil {
		log.Printf("Error fetching user %d data: %v", userID, err)
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}
	
	// 3. Получение API Key (для отображения в профиле)
	apiKey, err := repository.GetAPIKeyByUserID(userID)
	// Игнорируем ошибку, если у пользователя еще нет ключа.
	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Printf("Warning: Failed to fetch API Key for user %d: %v", userID, err)
	}

	// 4. Получение Grade пользователя
	gradeInfo, err := repository.GetUserGradeInfo(userID)
	if err != nil {
		log.Printf("Warning: Failed to fetch grade info for user %d: %v", userID, err)
		// Используем стандартный Grade если не удалось получить
		gradeInfo = &models.GradeInfo{
			Grade:      "STANDARD",
			TotalSpent: decimal.Zero,
			FeePercent: decimal.NewFromFloat(6.70),
		}
	}

	// 5. Формирование ответа
	response := struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Balance   string `json:"balance"`
		Status    string `json:"status"`
		APIKey    string `json:"api_key"`
		Grade     string `json:"grade"`
		FeePercent string `json:"fee_percent"`
	}{
		ID:        user.ID,
		Email:     user.Email,
		Balance:   user.BalanceRub.String(),
		Status:    user.Status,
		APIKey:    apiKey,
		Grade:     gradeInfo.Grade,
		FeePercent: gradeInfo.FeePercent.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}