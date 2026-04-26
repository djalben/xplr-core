package domain

import "time"

// TrustedDevice — запись о доверенном устройстве (cookie-токен хранится у клиента, на сервере только SHA256-хэш).
type TrustedDevice struct {
	ID         UUID       `db:"id"`
	UserID     UUID       `db:"user_id"`
	TokenHash  string     `db:"token_hash"`
	UserAgent  string     `db:"user_agent"`
	IP         *string    `db:"ip"`
	CreatedAt  time.Time  `db:"created_at"`
	LastUsedAt *time.Time `db:"last_used_at"`
	ExpiresAt  time.Time  `db:"expires_at"`
}
