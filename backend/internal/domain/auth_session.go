package domain

import "time"

// AuthSession — запись об успешном входе (для «Последняя активность»).
type AuthSession struct {
	ID        UUID      `db:"id"`
	UserID    UUID      `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	IP        *string   `db:"ip"`
	UserAgent string    `db:"user_agent"`
}
