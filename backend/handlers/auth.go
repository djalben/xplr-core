package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/utils"
	"github.com/shopspring/decimal"
)

// RegisterHandler - Регистрация нового пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация входных данных
	if strings.TrimSpace(req.Email) == "" {
		http.Error(w, "Email cannot be empty", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Хеширование пароля
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Создание пользователя
	user := models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Balance:      decimal.NewFromInt(0), // Инициализация баланса
		Status:       "ACTIVE",
	}

	createdUser, err := repository.CreateUser(user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		// Проверка на дубликат email
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Обработать реферальный код (если указан)
	if req.ReferralCode != "" {
		if err := repository.ProcessReferralRegistration(createdUser.ID, req.ReferralCode); err != nil {
			log.Printf("Warning: Failed to process referral code %s for user %d: %v", req.ReferralCode, createdUser.ID, err)
			// Не блокируем регистрацию, если реферальный код невалиден
		}
	}

	// Создать Grade для нового пользователя
	if _, err := repository.CreateUserGrade(createdUser.ID); err != nil {
		log.Printf("Warning: Failed to create user grade for user %d: %v", createdUser.ID, err)
		// Не блокируем регистрацию
	}

	// Генерация JWT токена (авто-логин после регистрации)
	token, err := utils.GenerateJWT(createdUser.ID)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	response := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":         createdUser.ID,
			"email":      createdUser.Email,
			"balance":    createdUser.BalanceRub.String(),
			"status":     "ACTIVE",
			"created_at": createdUser.CreatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// LoginHandler - Вход пользователя в систему
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация входных данных
	if strings.TrimSpace(req.Email) == "" {
		http.Error(w, "Email cannot be empty", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Password) == "" {
		http.Error(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	// Получение пользователя по email
	user, err := repository.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Login failed for email %s: %v", req.Email, err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		log.Printf("Invalid password for email %s", req.Email)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Генерация JWT токена
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("Error generating JWT for user %d: %v", user.ID, err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	response := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":      user.ID,
			"email":   user.Email,
			"balance": user.BalanceRub.String(),
			"status":  user.Status,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
