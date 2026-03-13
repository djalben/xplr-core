package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
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
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Генерация персонального реферального кода для нового пользователя
	refCode, err := repository.GetUserReferralCode(createdUser.ID)
	if err != nil {
		log.Printf("Warning: Failed to generate referral code for user %d: %v", createdUser.ID, err)
	} else {
		repository.SyncUserReferralCode(createdUser.ID, refCode)
	}

	// Обработать реферальный код (если указан)
	if req.ReferralCode != "" {
		if err := repository.ProcessReferralRegistration(createdUser.ID, req.ReferralCode); err != nil {
			log.Printf("Warning: Failed to process referral code %s for user %d: %v", req.ReferralCode, createdUser.ID, err)
		} else {
			// Также сохраняем referred_by в users таблицу
			referrerID := repository.GetReferrerID(createdUser.ID)
			if referrerID > 0 {
				if err := repository.SetReferredBy(createdUser.ID, referrerID); err != nil {
					log.Printf("Warning: Failed to set referred_by for user %d: %v", createdUser.ID, err)
				}
			}
		}
	}

	// Создать Grade для нового пользователя
	if _, err := repository.CreateUserGrade(createdUser.ID); err != nil {
		log.Printf("Warning: Failed to create user grade for user %d: %v", createdUser.ID, err)
		// Не блокируем регистрацию
	}

	// Генерация токена подтверждения email и отправка письма
	verifyToken, err := repository.CreateVerificationToken(createdUser.ID)
	if err != nil {
		log.Printf("Warning: Failed to create verification token for user %d: %v", createdUser.ID, err)
	} else {
		if err := service.SendVerificationEmail(createdUser.Email, verifyToken); err != nil {
			log.Printf("Warning: Failed to send verification email to %s: %v", createdUser.Email, err)
		}
	}

	// Отправка приветственного письма (async)
	go func(email string) {
		if err := service.SendWelcomeEmail(email); err != nil {
			log.Printf("Warning: Failed to send welcome email to %s: %v", email, err)
		}
	}(createdUser.Email)

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
			"id":          createdUser.ID,
			"email":       createdUser.Email,
			"balance":     createdUser.BalanceRub.String(),
			"status":      "ACTIVE",
			"is_verified": false,
			"created_at":  createdUser.CreatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// VerifyEmailHandler — GET /api/v1/auth/verify?token=...
// Подтверждает email пользователя по токену из письма.
func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token parameter", http.StatusBadRequest)
		return
	}

	userID, err := repository.VerifyToken(token)
	if err != nil {
		log.Printf("Email verification failed: %v", err)
		http.Error(w, "Verification failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Email successfully verified",
		"user_id": userID,
	})
}

// ResetPasswordRequestHandler — POST /api/v1/auth/reset-password-request
// Принимает email, создаёт токен сброса и отправляет письмо.
func ResetPasswordRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		http.Error(w, "Email cannot be empty", http.StatusBadRequest)
		return
	}

	// Всегда возвращаем 200 (не раскрываем существование email)
	user, err := repository.GetUserByEmail(email)
	if err != nil {
		log.Printf("[RESET] Reset requested for unknown email: %s", email)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "If this email is registered, a reset link has been sent.",
		})
		return
	}

	token, err := repository.CreatePasswordResetToken(user.ID)
	if err != nil {
		log.Printf("[RESET] Failed to create reset token for user %d: %v", user.ID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := service.SendPasswordResetEmail(user.Email, token); err != nil {
		log.Printf("[RESET] Failed to send reset email to %s: %v", user.Email, err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "If this email is registered, a reset link has been sent.",
	})
}

// ResetPasswordHandler — POST /api/v1/auth/reset-password
// Принимает токен и новый пароль, устанавливает новый пароль.
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	userID, err := repository.ValidatePasswordResetToken(req.Token)
	if err != nil {
		log.Printf("[RESET] Invalid reset token: %v", err)
		http.Error(w, "Invalid or expired reset link", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("[RESET] Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := repository.UpdateUserPassword(userID, hashedPassword); err != nil {
		log.Printf("[RESET] Failed to update password for user %d: %v", userID, err)
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	if err := repository.MarkPasswordResetTokenUsed(req.Token); err != nil {
		log.Printf("[RESET] Failed to mark token as used: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Password successfully reset",
	})
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
			"id":          user.ID,
			"email":       user.Email,
			"balance":     user.BalanceRub.String(),
			"status":      user.Status,
			"is_verified": user.IsVerified,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
