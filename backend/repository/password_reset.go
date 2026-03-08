package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// CreatePasswordResetToken — создаёт токен сброса пароля для пользователя.
// Токен действителен 1 час.
func CreatePasswordResetToken(userID int) (string, error) {
	// Генерируем случайный токен
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(bytes)

	expiresAt := time.Now().Add(1 * time.Hour)

	// Инвалидируем старые токены пользователя
	_, _ = GlobalDB.Exec(
		`UPDATE password_reset_tokens SET used = TRUE WHERE user_id = $1 AND used = FALSE`,
		userID,
	)

	// Создаём новый токен
	_, err := GlobalDB.Exec(
		`INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, token, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to save reset token: %w", err)
	}

	return token, nil
}

// ValidatePasswordResetToken — проверяет токен сброса пароля.
// Возвращает user_id если токен валиден.
func ValidatePasswordResetToken(token string) (int, error) {
	var userID int
	var expiresAt time.Time
	var used bool

	err := GlobalDB.QueryRow(
		`SELECT user_id, expires_at, used FROM password_reset_tokens WHERE token = $1`,
		token,
	).Scan(&userID, &expiresAt, &used)
	if err != nil {
		return 0, fmt.Errorf("invalid or expired token")
	}

	if used {
		return 0, fmt.Errorf("token already used")
	}

	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("token expired")
	}

	return userID, nil
}

// MarkPasswordResetTokenUsed — помечает токен как использованный.
func MarkPasswordResetTokenUsed(token string) error {
	_, err := GlobalDB.Exec(
		`UPDATE password_reset_tokens SET used = TRUE WHERE token = $1`,
		token,
	)
	return err
}

// UpdateUserPassword — обновляет хеш пароля пользователя.
func UpdateUserPassword(userID int, hashedPassword string) error {
	_, err := GlobalDB.Exec(
		`UPDATE users SET password_hash = $1 WHERE id = $2`,
		hashedPassword, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}
