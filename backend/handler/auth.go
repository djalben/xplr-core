package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/djalben/xplr-core/backend/domain"
	"github.com/djalben/xplr-core/backend/pkg/utils"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/djalben/xplr-core/backend/telegram"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shopspring/decimal"
)

// RegisterHandler - Регистрация нового пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
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
	user := domain.User{
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
			// Check if existing user is unverified — allow re-registration
			existing, lookupErr := repository.GetUserByEmail(req.Email)
			if lookupErr == nil && !existing.IsVerified {
				log.Printf("[REGISTER] Unverified user %d re-registering — updating password & resending verification link", existing.ID)
				// Update password
				if upErr := repository.UpdateUserPassword(existing.ID, hashedPassword); upErr != nil {
					log.Printf("[REGISTER] Failed to update password for user %d: %v", existing.ID, upErr)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				// Ensure tables exist (prod may have missed migrations)
				repository.RunSchemaGuard()

				// Create verification token + send email link
				token, tokErr := repository.CreateVerificationToken(existing.ID)
				if tokErr != nil {
					log.Printf("[REGISTER] Failed to create verification token for user %d: %v", existing.ID, tokErr)
					http.Error(w, "Failed to generate verification link", http.StatusInternalServerError)
					return
				}
				if sendErr := service.SendVerificationEmail(existing.Email, token); sendErr != nil {
					log.Printf("[REGISTER] Failed to send verification email to %s: %v", existing.Email, sendErr)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"email":        existing.Email,
					"message":      "Письмо подтверждения отправлено повторно. Откройте ссылку из письма.",
					"is_verified":  false,
				})
				return
			}
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

	// Ensure tables exist (prod may have missed migrations)
	repository.RunSchemaGuard()

	// Create verification token + send email link
	token, err := repository.CreateVerificationToken(createdUser.ID)
	if err != nil {
		log.Printf("[REGISTER] Failed to create verification token for user %d: %v", createdUser.ID, err)
		http.Error(w, "Failed to generate verification link", http.StatusInternalServerError)
		return
	}

	if err := service.SendVerificationEmail(createdUser.Email, token); err != nil {
		log.Printf("[REGISTER] Failed to send verification email to %s: %v", createdUser.Email, err)
		// Don't block registration — user can resend later
	}

	// Admin bootstrap: если в системе нет ни одного админа, первый пользователь становится админом
	if !repository.HasAnyAdmin() {
		log.Printf("[BOOTSTRAP] 🚀 No admins found — promoting first user %d (%s) to admin", createdUser.ID, createdUser.Email)
		if err := repository.PromoteToAdmin(createdUser.ID); err != nil {
			log.Printf("[BOOTSTRAP] Warning: failed to auto-promote: %v", err)
		}
	}

	// Уведомление админам о новом пользователе (async)
	go func(email string) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.Header.Get("X-Real-IP")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}
		msg := fmt.Sprintf(
			"🔥 <b>Новый пользователь!</b>\n\n"+
				"📧 <b>Email:</b> %s\n"+
				"🌍 <b>IP:</b> %s",
			email, ip,
		)
		telegram.NotifyAdmins(msg, "👤 Открыть в админке", "https://xplr.pro/admin/users")
	}(createdUser.Email)

	// NO TOKEN, NO AUTO-LOGIN — user must verify email by link, then login manually
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"email":        createdUser.Email,
		"is_verified":  false,
		"message":      "Регистрация успешна. Подтвердите email по ссылке из письма.",
	})
}

// VerifyEmailHandler — GET /api/v1/auth/verify-email?token=...
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

	log.Printf("[VERIFY] ✅ Email verified for user %d", userID)
	// Redirect to /auth with verified flag — user must log in manually (no auto-session)
	appDomain := os.Getenv("APP_DOMAIN")
	if appDomain == "" {
		appDomain = "https://xplr.pro"
	}
	http.Redirect(w, r, appDomain+"/auth?verified=1", http.StatusFound)
}

// ResendVerificationHandler — POST /api/v1/auth/resend-verification
// Повторно отправляет письмо подтверждения email (публичный, по email).
func ResendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Always return 200 (don't leak whether email exists)
	w.Header().Set("Content-Type", "application/json")

	user, err := repository.GetUserByEmail(email)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "If this email is registered, a verification link has been sent."})
		return
	}

	// Already verified — no need to resend
	if user.IsVerified {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "If this email is registered, a verification link has been sent."})
		return
	}

	token, err := repository.CreateVerificationToken(user.ID)
	if err != nil {
		log.Printf("[RESEND-VERIFY] Failed to create token for %s: %v", email, err)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "If this email is registered, a verification link has been sent."})
		return
	}

	if err := service.SendVerificationEmail(email, token); err != nil {
		log.Printf("[RESEND-VERIFY] Failed to send email to %s: %v", email, err)
	}

	log.Printf("[RESEND-VERIFY] Verification email resent to %s", email)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "If this email is registered, a verification link has been sent."})
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
// Self-healing: при ошибках БД пробует минимальный запрос, авто-создаёт недостающие данные.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация входных данных
	email := strings.TrimSpace(req.Email)
	if email == "" {
		http.Error(w, "Email cannot be empty", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		http.Error(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	// ── Шаг 1: попробовать полный запрос ──
	user, err := repository.GetUserByEmail(email)
	usedFallback := false

	if err != nil {
		errMsg := err.Error()
		isNotFound := strings.Contains(errMsg, "не найден") || strings.Contains(errMsg, "not found")

		if isNotFound {
			log.Printf("[LOGIN] User not found: %s", email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// ── Шаг 1b: DB error → self-heal schema, then retry full query ──
		log.Printf("[LOGIN] ⚠️  Full query failed for %s: %v — running schema guard", email, err)
		repository.RunSchemaGuard()

		// Retry full query after heal
		user, err = repository.GetUserByEmail(email)
		if err != nil {
			// Still failing → fallback to basic
			log.Printf("[LOGIN] ⚠️  Retry failed, using basic fallback for %s: %v", email, err)
			user, err = repository.GetUserByEmailBasic(email)
			if err != nil {
				if strings.Contains(err.Error(), "не найден") || strings.Contains(err.Error(), "not found") {
					log.Printf("[LOGIN] User not found (fallback): %s", email)
				} else {
					log.Printf("[LOGIN] ❌ CRITICAL DB ERROR for %s: %v", email, err)
				}
				http.Error(w, "Invalid email or password", http.StatusUnauthorized)
				return
			}
			usedFallback = true
		}
	}

	// ── Шаг 2: проверка пароля ──
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		log.Printf("[LOGIN] Wrong password for %s (user_id=%d)", email, user.ID)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// ── Шаг 2b: Email verification gate ──
	// Only blocks users who have NOT confirmed their email.
	// Existing confirmed users pass through normally.
	if !user.IsVerified {
		log.Printf("[LOGIN] ❌ Email not verified for %s (user_id=%d)", email, user.ID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":              "email_not_verified",
			"message":            "Пожалуйста, подтвердите ваш email. Проверьте почту.",
			"email_not_verified": true,
		})
		return
	}

	// ── Шаг 3: Авто-создание Grade если отсутствует ──
	if _, gradeErr := repository.GetUserGradeInfo(user.ID); gradeErr != nil {
		log.Printf("[LOGIN] Auto-creating grade for user %d", user.ID)
		if _, err := repository.CreateUserGrade(user.ID); err != nil {
			log.Printf("[LOGIN] Warning: failed to auto-create grade for user %d: %v", user.ID, err)
		}
	}

	// ── Шаг 4: Admin bootstrap — если нет ни одного админа, назначаем текущего ──
	if !repository.HasAnyAdmin() {
		log.Printf("[BOOTSTRAP] 🚀 No admins in system — promoting user %d (%s) to admin", user.ID, user.Email)
		if err := repository.PromoteToAdmin(user.ID); err != nil {
			log.Printf("[BOOTSTRAP] Warning: auto-promote failed: %v", err)
		} else {
			user.IsAdmin = true
			user.Role = "admin"
		}
	}

	// ── Шаг 5: If used fallback, re-read full user to get is_admin/role ──
	if usedFallback {
		if fullUser, err := repository.GetUserByID(user.ID); err == nil {
			user = fullUser
		}
	}

	// ── Шаг 5b: 2FA gate — if enabled and device not trusted, return half_auth_token ──
	_, twoFAEnabled, _ := repository.GetTwoFactorSecret(user.ID)
	if twoFAEnabled {
		// Check trusted device cookie
		deviceTrusted := false
		if cookie, cookieErr := r.Cookie("xplr_trusted_device"); cookieErr == nil && cookie.Value != "" {
			deviceTrusted = repository.IsTrustedDevice(user.ID, cookie.Value)
		}
		if !deviceTrusted {
			halfToken, htErr := utils.GenerateHalfAuthJWT(user.ID)
			if htErr != nil {
				log.Printf("[LOGIN] ❌ Half-auth JWT failed for user %d: %v", user.ID, htErr)
				http.Error(w, "Failed to generate token", http.StatusInternalServerError)
				return
			}
			log.Printf("[LOGIN] 🔐 2FA required for %s (user_id=%d)", user.Email, user.ID)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"requires_2fa":    true,
				"half_auth_token": halfToken,
			})
			return
		}
		log.Printf("[LOGIN] ✅ 2FA bypassed (trusted device) for %s", user.Email)
	}

	// ── Шаг 6: генерация JWT с ролевой информацией + token_version ──
	isAdmin := user.IsAdmin || user.Role == "admin"
	role := user.Role
	if role == "" {
		role = "user"
	}
	tv, _ := repository.GetTokenVersion(user.ID)
	token, err := utils.GenerateJWT(user.ID, isAdmin, role, tv)
	if err != nil {
		log.Printf("[LOGIN] ❌ JWT generation failed for user %d: %v", user.ID, err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("[LOGIN] ✅ Success: %s (user_id=%d, is_admin=%v, role=%s, fallback=%v)",
		user.Email, user.ID, isAdmin, role, usedFallback)

	// ── Шаг 7: создаём запись о сессии (IP + User-Agent) ──
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	device := r.Header.Get("User-Agent")
	if len(device) > 200 {
		device = device[:200]
	}

	// ── Шаг 7b: Проверка нового IP — уведомление о входе с нового устройства ──
	if repository.IsNewIPForUser(user.ID, ip) {
		log.Printf("[SECURITY] New IP detected for user %d (%s): ip=%s", user.ID, user.Email, ip)
		go service.NotifyUser(user.ID, "Вход с нового устройства",
			fmt.Sprintf("⚠️ <b>Вход в аккаунт с нового устройства/IP</b>\n\n"+
				"IP: <b>%s</b>\n"+
				"Устройство: <b>%s</b>\n\n"+
				"Если это были не вы, немедленно смените пароль.\n\n"+
				"<a href=\"https://xplr.pro/settings\">Настройки безопасности</a>",
				ip, device))
	}

	if err := repository.CreateUserSession(user.ID, ip, device); err != nil {
		log.Printf("[LOGIN] Warning: failed to create session for user %d: %v", user.ID, err)
	}

	// Успешный ответ — включает is_admin и role для фронтенда
	response := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":                 user.ID,
			"email":              user.Email,
			"balance":            user.BalanceRub.String(),
			"status":             user.Status,
			"is_verified":        user.IsVerified,
			"is_admin":           isAdmin,
			"role":               role,
			"two_factor_enabled": twoFAEnabled,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RefreshTokenHandler — POST /api/v1/auth/refresh-token
// Re-reads user from DB and issues a fresh JWT with current is_admin/role.
// Requires valid existing JWT in Authorization header.
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse existing token to get user_id
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return utils.GetJWTSecret(), nil
	})
	if err != nil || !parsedToken.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid claims", http.StatusUnauthorized)
		return
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "Invalid user_id in token", http.StatusUnauthorized)
		return
	}

	userID := int(userIDFloat)

	// Re-read user from DB (fresh data)
	user, err := repository.GetUserByID(userID)
	if err != nil {
		user, err = repository.GetUserByIDBasic(userID)
		if err != nil {
			log.Printf("[REFRESH] User %d not found: %v", userID, err)
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
	}

	isAdmin := user.IsAdmin || user.Role == "admin"
	role := user.Role
	if role == "" {
		role = "user"
	}

	// Check token_version — reject stale tokens
	tv, _ := repository.GetTokenVersion(user.ID)
	if tvClaim, hasTv := claims["tv"].(float64); hasTv {
		if int(tvClaim) < tv {
			log.Printf("[REFRESH] ❌ Stale token for user %d (token_tv=%d, db_tv=%d)", userID, int(tvClaim), tv)
			http.Error(w, "Session expired. Please login again.", http.StatusUnauthorized)
			return
		}
	}

	newToken, err := utils.GenerateJWT(user.ID, isAdmin, role, tv)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("[REFRESH] ✅ Token refreshed for %s (is_admin=%v, role=%s)", user.Email, isAdmin, role)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": newToken,
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"is_admin": isAdmin,
			"role":     role,
		},
	})
}
