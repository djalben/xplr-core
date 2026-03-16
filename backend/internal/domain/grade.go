package domain

import "time"

type UserGrade struct {
	ID         UUID      `json:"id" db:"id"`
	UserID     UUID      `json:"userId" db:"user_id"`
	Grade      string    `json:"grade" db:"grade"`
	TotalSpent Numeric   `json:"totalSpent" db:"total_spent"`
	FeePercent Numeric   `json:"feePercent" db:"fee_percent"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}
