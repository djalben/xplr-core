package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"
)

// CreateVerificationToken — генерирует уникальный токен подтверждения email и сохраняет в БД.
// Токен действителен 24 часа.
func CreateVerificationToken(userID int) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	// Генерируем криптографически безопасный токен
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(bytes)

	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := GlobalDB.Exec(
		`INSERT INTO verification_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, token, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to save verification token: %w", err)
	}

	log.Printf("Verification token created for user %d, expires at %s", userID, expiresAt.Format(time.RFC3339))
	return token, nil
}

// VerifyToken — проверяет токен и помечает пользователя как подтверждённого.
// Возвращает user_id при успехе.
func VerifyToken(token string) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	tx, err := GlobalDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var userID int
	var used bool
	var expiresAt time.Time

	err = tx.QueryRow(
		`SELECT user_id, used, expires_at FROM verification_tokens WHERE token = $1`,
		token,
	).Scan(&userID, &used, &expiresAt)
	if err != nil {
		return 0, fmt.Errorf("invalid verification token")
	}

	if used {
		return 0, fmt.Errorf("token already used")
	}

	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("token expired")
	}

	// Помечаем токен как использованный
	_, err = tx.Exec(`UPDATE verification_tokens SET used = TRUE WHERE token = $1`, token)
	if err != nil {
		return 0, fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Помечаем пользователя как подтверждённого
	_, err = tx.Exec(`UPDATE users SET is_verified = TRUE WHERE id = $1`, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to verify user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit: %w", err)
	}

	log.Printf("✅ User %d email verified successfully", userID)
	return userID, nil
}

// CreateOTPCode — generates a 6-digit OTP code for email verification.
// Stored in verification_tokens table. Expires in 10 minutes.
// Invalidates any previous OTP for the same user.
func CreateOTPCode(userID int) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	// Invalidate previous OTPs for this user
	_, _ = GlobalDB.Exec(`UPDATE verification_tokens SET used = TRUE WHERE user_id = $1 AND used = FALSE`, userID)

	// Generate 6-digit code
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	code := fmt.Sprintf("%06d", n.Int64())

	expiresAt := time.Now().Add(10 * time.Minute)

	_, err = GlobalDB.Exec(
		`INSERT INTO verification_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, code, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to save OTP code: %w", err)
	}

	log.Printf("[OTP] Created code for user %d, expires at %s", userID, expiresAt.Format(time.RFC3339))
	return code, nil
}

// VerifyOTPCode — verifies a 6-digit OTP code and marks user as verified.
// Returns user_id on success.
func VerifyOTPCode(email, code string) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	tx, err := GlobalDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var userID int
	var used bool
	var expiresAt time.Time

	err = tx.QueryRow(
		`SELECT vt.user_id, vt.used, vt.expires_at
		 FROM verification_tokens vt
		 JOIN users u ON u.id = vt.user_id
		 WHERE u.email = $1 AND vt.token = $2
		 ORDER BY vt.created_at DESC LIMIT 1`,
		email, code,
	).Scan(&userID, &used, &expiresAt)
	if err != nil {
		return 0, fmt.Errorf("invalid OTP code")
	}

	if used {
		return 0, fmt.Errorf("code already used")
	}

	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("code expired")
	}

	// Mark token as used
	_, err = tx.Exec(`UPDATE verification_tokens SET used = TRUE WHERE user_id = $1 AND token = $2`, userID, code)
	if err != nil {
		return 0, fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	// Mark user as verified
	_, err = tx.Exec(`UPDATE users SET is_verified = TRUE WHERE id = $1`, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to verify user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit: %w", err)
	}

	log.Printf("[OTP] ✅ User %d (email=%s) verified via OTP", userID, email)
	return userID, nil
}

// IsUserVerified — проверяет, подтверждён ли email пользователя.
func IsUserVerified(userID int) (bool, error) {
	if GlobalDB == nil {
		return false, fmt.Errorf("database connection not initialized")
	}

	var verified bool
	err := GlobalDB.QueryRow(
		`SELECT COALESCE(is_verified, FALSE) FROM users WHERE id = $1`, userID,
	).Scan(&verified)
	if err != nil {
		return false, fmt.Errorf("failed to check verification status: %w", err)
	}
	return verified, nil
}
