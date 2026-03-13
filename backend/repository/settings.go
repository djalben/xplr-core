package repository

import (
	"fmt"
	"log"
	"time"
)

// ── User Sessions ──

type UserSession struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	IP         string `json:"ip"`
	Location   string `json:"location"`
	Device     string `json:"device"`
	LastActive string `json:"last_active"`
	CreatedAt  string `json:"created_at"`
}

func CreateUserSession(userID int, ip, device string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`INSERT INTO user_sessions (user_id, ip, device, location) VALUES ($1, $2, $3, '')`,
		userID, ip, device,
	)
	return err
}

func GetRecentSessions(userID int, limit int) ([]UserSession, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	rows, err := GlobalDB.Query(
		`SELECT id, user_id, COALESCE(ip,''), COALESCE(location,''), COALESCE(device,''), last_active, created_at
		 FROM user_sessions WHERE user_id = $1 ORDER BY last_active DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []UserSession
	for rows.Next() {
		var s UserSession
		var lastActive, createdAt time.Time
		if err := rows.Scan(&s.ID, &s.UserID, &s.IP, &s.Location, &s.Device, &lastActive, &createdAt); err != nil {
			continue
		}
		s.LastActive = lastActive.Format(time.RFC3339)
		s.CreatedAt = createdAt.Format(time.RFC3339)
		sessions = append(sessions, s)
	}
	if sessions == nil {
		sessions = []UserSession{}
	}
	return sessions, nil
}

func DeleteAllUserSessions(userID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(`DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}

// ── Change Password ──

func UpdatePasswordHash(userID int, newHash string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(`UPDATE users SET password_hash = $1 WHERE id = $2`, newHash, userID)
	return err
}

func GetPasswordHash(userID int) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf("database connection not initialized")
	}
	var hash string
	err := GlobalDB.QueryRow(`SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&hash)
	return hash, err
}

// ── 2FA (TOTP) ──

func SetTwoFactorSecret(userID int, secret string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE users SET two_factor_secret = $1 WHERE id = $2`,
		secret, userID,
	)
	return err
}

func EnableTwoFactor(userID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE users SET two_factor_enabled = TRUE WHERE id = $1`, userID,
	)
	return err
}

func DisableTwoFactor(userID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE users SET two_factor_enabled = FALSE, two_factor_secret = NULL WHERE id = $1`, userID,
	)
	return err
}

func GetTwoFactorSecret(userID int) (string, bool, error) {
	if GlobalDB == nil {
		return "", false, fmt.Errorf("database connection not initialized")
	}
	var secret *string
	var enabled bool
	err := GlobalDB.QueryRow(
		`SELECT two_factor_secret, COALESCE(two_factor_enabled, FALSE) FROM users WHERE id = $1`, userID,
	).Scan(&secret, &enabled)
	if err != nil {
		return "", false, err
	}
	if secret == nil {
		return "", enabled, nil
	}
	return *secret, enabled, nil
}

// ── Notification Preferences ──

type NotificationPrefs struct {
	NotifyTransactions bool `json:"notify_transactions"`
	NotifyBalance      bool `json:"notify_balance"`
	NotifySecurity     bool `json:"notify_security"`
}

func GetNotificationPrefs(userID int) (NotificationPrefs, error) {
	p := NotificationPrefs{NotifyTransactions: true, NotifyBalance: true, NotifySecurity: true}
	if GlobalDB == nil {
		return p, fmt.Errorf("database connection not initialized")
	}
	err := GlobalDB.QueryRow(
		`SELECT COALESCE(notify_transactions, TRUE), COALESCE(notify_balance, TRUE), COALESCE(notify_security, TRUE) FROM users WHERE id = $1`,
		userID,
	).Scan(&p.NotifyTransactions, &p.NotifyBalance, &p.NotifySecurity)
	return p, err
}

func UpdateNotificationPrefs(userID int, p NotificationPrefs) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE users SET notify_transactions = $1, notify_balance = $2, notify_security = $3 WHERE id = $4`,
		p.NotifyTransactions, p.NotifyBalance, p.NotifySecurity, userID,
	)
	return err
}

// ── Display Name ──

func UpdateDisplayName(userID int, name string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(`UPDATE users SET display_name = $1 WHERE id = $2`, name, userID)
	return err
}

// ── Extended /me data ──

type MeExtended struct {
	ID                 int    `json:"id"`
	Email              string `json:"email"`
	DisplayName        string `json:"display_name"`
	IsVerified         bool   `json:"is_verified"`
	VerificationStatus string `json:"verification_status"`
	TwoFactorEnabled   bool   `json:"two_factor_enabled"`
	TelegramLinked     bool   `json:"telegram_linked"`
	Role               string `json:"role"`
	IsAdmin            bool   `json:"is_admin"`
}

func GetMeExtended(userID int) (MeExtended, error) {
	var m MeExtended
	if GlobalDB == nil {
		return m, fmt.Errorf("database connection not initialized")
	}
	err := GlobalDB.QueryRow(
		`SELECT id, email, COALESCE(display_name, ''), COALESCE(is_verified, FALSE),
		        COALESCE(verification_status, 'pending'), COALESCE(two_factor_enabled, FALSE),
		        (telegram_chat_id IS NOT NULL), COALESCE(role, 'user'), COALESCE(is_admin, FALSE)
		 FROM users WHERE id = $1`, userID,
	).Scan(&m.ID, &m.Email, &m.DisplayName, &m.IsVerified, &m.VerificationStatus,
		&m.TwoFactorEnabled, &m.TelegramLinked, &m.Role, &m.IsAdmin)
	if err != nil {
		log.Printf("[SETTINGS] GetMeExtended failed for user %d: %v", userID, err)
	}
	return m, err
}
