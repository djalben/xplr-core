package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/djalben/xplr-core/backend/pkg/utils"
	"github.com/djalben/xplr-core/backend/repository"
)

// LoginVerify2FAHandler — POST /api/v1/auth/2fa/verify
// Validates half_auth_token + TOTP code (or recovery code).
// If "remember_device" is true, sets HttpOnly cookie xplr_trusted_device (30 days).
// Returns full Access token on success.
func LoginVerify2FAHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		HalfAuthToken  string `json:"half_auth_token"`
		Code           string `json:"code"`
		RememberDevice bool   `json:"remember_device"`
		Fingerprint    string `json:"fingerprint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate half-auth token
	userID, err := utils.ParseHalfAuthJWT(req.HalfAuthToken)
	if err != nil {
		log.Printf("[2FA-VERIFY] Invalid half-auth token: %v", err)
		http.Error(w, "Invalid or expired verification session", http.StatusUnauthorized)
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		http.Error(w, "Code is required", http.StatusBadRequest)
		return
	}

	// Get 2FA secret
	secret, enabled, err := repository.GetTwoFactorSecret(userID)
	if err != nil || !enabled || secret == "" {
		http.Error(w, "2FA not enabled for this account", http.StatusBadRequest)
		return
	}

	verified := false

	// Try TOTP first (6-digit code)
	if len(code) == 6 {
		if verifyTOTP(secret, code) {
			verified = true
		}
	}

	// Try recovery code if TOTP didn't match
	if !verified {
		if tryRecoveryCode(userID, code) {
			verified = true
			log.Printf("[2FA-VERIFY] Recovery code used for user %d", userID)
		}
	}

	if !verified {
		log.Printf("[2FA-VERIFY] ❌ Invalid code for user %d", userID)
		http.Error(w, "Invalid verification code", http.StatusForbidden)
		return
	}

	// Fetch full user for JWT generation
	user, err := repository.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	isAdmin := user.IsAdmin || user.Role == "admin"
	role := user.Role
	if role == "" {
		role = "user"
	}

	token, err := utils.GenerateJWT(user.ID, isAdmin, role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set trusted device cookie if requested
	if req.RememberDevice {
		ua := r.Header.Get("User-Agent")
		fp := req.Fingerprint
		if fp == "" {
			fp = ua
		}
		deviceHash := hashDeviceFingerprint(ua, fp)
		if err := repository.AddTrustedDevice(userID, deviceHash); err != nil {
			log.Printf("[2FA-VERIFY] Warning: failed to save trusted device for user %d: %v", userID, err)
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "xplr_trusted_device",
			Value:    deviceHash,
			Path:     "/",
			MaxAge:   30 * 24 * 3600, // 30 days
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		log.Printf("[2FA-VERIFY] Trusted device saved for user %d", userID)
	}

	// Create session
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
	_ = repository.CreateUserSession(userID, ip, device)

	log.Printf("[2FA-VERIFY] ✅ 2FA verified for user %d (%s)", userID, user.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":          user.ID,
			"email":       user.Email,
			"balance":     user.BalanceRub.String(),
			"status":      user.Status,
			"is_verified": user.IsVerified,
			"is_admin":    isAdmin,
			"role":        role,
		},
	})
}

// tryRecoveryCode checks if the provided code matches any hashed recovery code.
// Burns the code on match.
func tryRecoveryCode(userID int, code string) bool {
	codes, err := repository.GetRecoveryCodes(userID)
	if err != nil || len(codes) == 0 {
		return false
	}
	codeHash := hashRecoveryCode(code)
	for i, stored := range codes {
		if stored == codeHash {
			if err := repository.BurnRecoveryCode(userID, i); err != nil {
				log.Printf("[2FA] Warning: failed to burn recovery code for user %d: %v", userID, err)
			}
			return true
		}
	}
	return false
}

// hashRecoveryCode hashes a recovery code with SHA-256.
func hashRecoveryCode(code string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(strings.ToUpper(code))))
	return hex.EncodeToString(h[:])
}

// hashDeviceFingerprint creates a hash from User-Agent + fingerprint.
func hashDeviceFingerprint(userAgent, fingerprint string) string {
	data := fmt.Sprintf("%s|%s", userAgent, fingerprint)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// generateRecoveryCodes creates 5 random 8-char alphanumeric codes.
// Returns both plain codes (to show user once) and their hashes (to store).
func generateRecoveryCodes() (plain []string, hashed []string) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // No O/0/I/1 for readability
	for i := 0; i < 5; i++ {
		code := make([]byte, 8)
		for j := range code {
			code[j] = chars[cryptoRandInt(len(chars))]
		}
		codeStr := string(code)
		plain = append(plain, codeStr)
		hashed = append(hashed, hashRecoveryCode(codeStr))
	}
	return
}
