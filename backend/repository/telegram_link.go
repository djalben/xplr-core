package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

// StoreTelegramLinkCode generates a random UUID-like code, stores it in the DB
// linked to the userID, and returns the code. Expires in 15 minutes.
func StoreTelegramLinkCode(userID int) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	// Generate a random 16-byte hex code
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate link code: %w", err)
	}
	code := hex.EncodeToString(b)

	// Upsert: one active code per user at a time
	_, err := GlobalDB.Exec(
		`INSERT INTO telegram_link_codes (user_id, code, expires_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET code = $2, expires_at = $3`,
		userID, code, time.Now().Add(15*time.Minute),
	)
	if err != nil {
		return "", fmt.Errorf("failed to store link code: %w", err)
	}
	log.Printf("[TELEGRAM] Link code generated for user %d", userID)
	return code, nil
}

// LookupTelegramLinkCode finds the userID associated with a non-expired code.
// Returns 0 if code is invalid or expired.
func LookupTelegramLinkCode(code string) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	var userID int
	err := GlobalDB.QueryRow(
		`SELECT user_id FROM telegram_link_codes WHERE code = $1 AND expires_at > NOW()`,
		code,
	).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// DeleteTelegramLinkCode removes a used code.
func DeleteTelegramLinkCode(code string) {
	if GlobalDB == nil {
		return
	}
	_, _ = GlobalDB.Exec(`DELETE FROM telegram_link_codes WHERE code = $1`, code)
}

// GetUserIDByChatID returns the user ID linked to a given Telegram chat_id.
// Returns 0 if no user is linked.
func GetUserIDByChatID(chatID int64) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}
	var userID int
	err := GlobalDB.QueryRow(
		`SELECT id FROM users WHERE telegram_chat_id = $1`, chatID,
	).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// GetUserEmailByChatID returns the email of the user linked to a given Telegram chat_id.
// Returns empty string if no user is linked.
func GetUserEmailByChatID(chatID int64) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf("database connection not initialized")
	}
	var email string
	err := GlobalDB.QueryRow(
		`SELECT email FROM users WHERE telegram_chat_id = $1`, chatID,
	).Scan(&email)
	if err != nil {
		return "", err
	}
	return email, nil
}

// UpdateTelegramChatIDInt64 updates the telegram_chat_id for a user (int64 version).
func UpdateTelegramChatIDInt64(userID int, chatID int64) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE users SET telegram_chat_id = $1 WHERE id = $2`,
		chatID, userID,
	)
	if err != nil {
		log.Printf("[TELEGRAM] Error updating chat_id for user %d: %v", userID, err)
		return err
	}
	log.Printf("[TELEGRAM] Chat ID %d linked to user %d", chatID, userID)
	return nil
}
