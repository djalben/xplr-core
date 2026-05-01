package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
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
