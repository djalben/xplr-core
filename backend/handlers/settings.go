package handlers

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/djalben/xplr-core/backend/telegram"
	"golang.org/x/crypto/bcrypt"
)

// cryptoRandInt returns a cryptographically random int in [0, max)
func cryptoRandInt(max int) int {
	var b [8]byte
	_, _ = rand.Read(b[:])
	n := int(binary.BigEndian.Uint64(b[:]) % uint64(max))
	return n
}

// ── GET /api/v1/user/settings/profile ──
func GetSettingsProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	me, err := repository.GetMeExtended(userID)
	if err != nil {
		http.Error(w, "Failed to fetch profile", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(me)
}

// ── PATCH /api/v1/user/settings/profile ──
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(req.DisplayName)
	if name == "" {
		http.Error(w, "Display name cannot be empty", http.StatusBadRequest)
		return
	}
	if err := repository.UpdateDisplayName(userID, name); err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"display_name": name})
}

// ── POST /api/v1/user/settings/change-password ──
func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	if len(req.NewPassword) < 8 {
		http.Error(w, "New password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Verify old password
	hash, err := repository.GetPasswordHash(userID)
	if err != nil {
		http.Error(w, "Failed to verify password", http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.OldPassword)); err != nil {
		http.Error(w, "Incorrect current password", http.StatusForbidden)
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	if err := repository.UpdatePasswordHash(userID, string(newHash)); err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	log.Printf("[SETTINGS] Password changed for user %d", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
}

// ── GET /api/v1/user/settings/sessions ──
func GetSessionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	sessions, err := repository.GetRecentSessions(userID, 5)
	if err != nil {
		http.Error(w, "Failed to fetch sessions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// ── POST /api/v1/user/settings/logout-all ──
func LogoutAllSessionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := repository.DeleteAllUserSessions(userID); err != nil {
		http.Error(w, "Failed to logout all sessions", http.StatusInternalServerError)
		return
	}
	log.Printf("[SETTINGS] User %d logged out from all devices", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "All sessions terminated"})
}

// ── GET /api/v1/user/settings/notifications ──
func GetNotificationPrefsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	prefs, err := repository.GetNotificationPrefs(userID)
	if err != nil {
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}
	notifPref := repository.GetNotificationPref(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"notify_transactions": prefs.NotifyTransactions,
		"notify_balance":      prefs.NotifyBalance,
		"notify_security":     prefs.NotifySecurity,
		"notification_pref":   notifPref,
	})
}

// ── PATCH /api/v1/user/settings/notifications ──
func UpdateNotificationPrefsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		NotifyTransactions *bool   `json:"notify_transactions,omitempty"`
		NotifyBalance      *bool   `json:"notify_balance,omitempty"`
		NotifySecurity     *bool   `json:"notify_security,omitempty"`
		NotificationPref   *string `json:"notification_pref,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	// Update notification_pref channel if provided
	if req.NotificationPref != nil {
		pref := *req.NotificationPref
		if err := repository.SetNotificationPref(userID, pref); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Update toggle prefs
	prefs, _ := repository.GetNotificationPrefs(userID)
	if req.NotifyTransactions != nil {
		prefs.NotifyTransactions = *req.NotifyTransactions
	}
	if req.NotifyBalance != nil {
		prefs.NotifyBalance = *req.NotifyBalance
	}
	if req.NotifySecurity != nil {
		prefs.NotifySecurity = *req.NotifySecurity
	}
	if err := repository.UpdateNotificationPrefs(userID, prefs); err != nil {
		http.Error(w, "Failed to update notifications", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ── POST /api/v1/user/settings/telegram/unlink ──
func UnlinkTelegramHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Ensure user won't be left with no notification channel
	pref := repository.GetNotificationPref(userID)
	if pref == "telegram" {
		// Switch to email before unlinking
		_ = repository.SetNotificationPref(userID, "email")
	}
	if err := repository.UnlinkTelegram(userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Telegram отвязан"})
}

// ── POST /api/v1/user/settings/2fa/unlink ──
func Unlink2FAHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := repository.DisableTwoFactor(userID); err != nil {
		http.Error(w, "Failed to disable 2FA", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "2FA отключена"})
}

// ── POST /api/v1/user/settings/2fa/setup ──
func Setup2FAHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate random secret
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		http.Error(w, "Failed to generate secret", http.StatusInternalServerError)
		return
	}
	secretB32 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)

	// Store secret (not yet enabled)
	if err := repository.SetTwoFactorSecret(userID, secretB32); err != nil {
		http.Error(w, "Failed to save secret", http.StatusInternalServerError)
		return
	}

	// Get user email for the TOTP URI
	me, err := repository.GetMeExtended(userID)
	if err != nil {
		http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		return
	}

	// otpauth URI for Google Authenticator
	otpURI := "otpauth://totp/XPLR:" + me.Email + "?secret=" + secretB32 + "&issuer=XPLR&digits=6&period=30"

	log.Printf("[2FA] Setup initiated for user %d", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"secret":  secretB32,
		"otp_uri": otpURI,
	})
}

// ── POST /api/v1/user/settings/2fa/verify ──
func Verify2FAHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	code := strings.TrimSpace(req.Code)
	if len(code) != 6 {
		http.Error(w, "Code must be 6 digits", http.StatusBadRequest)
		return
	}

	secret, _, err := repository.GetTwoFactorSecret(userID)
	if err != nil || secret == "" {
		http.Error(w, "2FA not set up", http.StatusBadRequest)
		return
	}

	// Verify TOTP code
	if !verifyTOTP(secret, code) {
		http.Error(w, "Invalid code", http.StatusForbidden)
		return
	}

	if err := repository.EnableTwoFactor(userID); err != nil {
		http.Error(w, "Failed to enable 2FA", http.StatusInternalServerError)
		return
	}

	log.Printf("[2FA] ✅ Enabled for user %d", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"two_factor_enabled": true})
}

// ── POST /api/v1/user/settings/2fa/disable ──
func Disable2FAHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := repository.DisableTwoFactor(userID); err != nil {
		http.Error(w, "Failed to disable 2FA", http.StatusInternalServerError)
		return
	}
	log.Printf("[2FA] Disabled for user %d", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"two_factor_enabled": false})
}

// ── POST /api/v1/user/settings/verify-email-request ──
func RequestEmailVerifyHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email, err := repository.GetUserEmail(userID)
	if err != nil || email == "" {
		http.Error(w, "Failed to fetch user email", http.StatusInternalServerError)
		return
	}

	// Generate 6-digit code
	code := fmt.Sprintf("%06d", cryptoRandInt(1000000))

	if err := repository.SetEmailVerifyCode(userID, code); err != nil {
		log.Printf("[EMAIL-VERIFY] Failed to store code for user %d: %v", userID, err)
		http.Error(w, "Failed to generate code", http.StatusInternalServerError)
		return
	}

	// Send code via Zoho SMTP
	if err := service.SendEmailVerifyCode(email, code); err != nil {
		log.Printf("[EMAIL-VERIFY] Failed to send code to %s: %v", email, err)
		http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
		return
	}

	log.Printf("[EMAIL-VERIFY] Code sent to %s for user %d", email, userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Verification code sent"})
}

// ── POST /api/v1/user/settings/verify-email-confirm ──
func ConfirmEmailVerifyHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	code := strings.TrimSpace(req.Code)
	if len(code) != 6 {
		http.Error(w, "Code must be 6 digits", http.StatusBadRequest)
		return
	}

	valid, err := repository.CheckEmailVerifyCode(userID, code)
	if err != nil {
		http.Error(w, "Verification failed", http.StatusInternalServerError)
		return
	}
	if !valid {
		http.Error(w, "Invalid or expired code", http.StatusForbidden)
		return
	}

	if err := repository.MarkEmailVerified(userID); err != nil {
		http.Error(w, "Failed to verify email", http.StatusInternalServerError)
		return
	}

	log.Printf("[EMAIL-VERIFY] ✅ Email verified for user %d", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"is_verified": true})
}

// ── POST /api/v1/user/settings/kyc ──
func SubmitKYCHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		Country   string `json:"country"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		BirthDate string `json:"birth_date"`
		Address   string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Country) == "" || strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		http.Error(w, "Country, first name and last name are required", http.StatusBadRequest)
		return
	}

	id, err := repository.CreateKYCRequest(userID, req.Country, req.FirstName, req.LastName, req.BirthDate, req.Address)
	if err != nil {
		log.Printf("[KYC] Failed to create request for user %d: %v", userID, err)
		http.Error(w, "Failed to submit KYC request", http.StatusInternalServerError)
		return
	}

	log.Printf("[KYC] ✅ Request %d submitted by user %d", id, userID)

	// Уведомление админам о новой заявке KYC (async)
	go func() {
		email := ""
		if u, err := repository.GetUserByID(userID); err == nil {
			email = u.Email
		}
		if email == "" {
			email = fmt.Sprintf("User #%d", userID)
		}
		msg := fmt.Sprintf(
			"📄 <b>Новая заявка на KYC!</b>\n\n"+
				"👤 <b>Пользователь:</b> %s\n"+
				"📋 <b>Статус:</b> Ожидает проверки",
			email,
		)
		telegram.NotifyAdmins(msg, "🔍 Проверить документы", "https://xplr.pro/admin/kyc")
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     id,
		"status": "pending",
	})
}

// ── GET /api/v1/user/settings/kyc ──
func GetKYCHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	kyc, err := repository.GetLatestKYCRequest(userID)
	if err != nil {
		// No KYC request found — return empty
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kyc)
}
