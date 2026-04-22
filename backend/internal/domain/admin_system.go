package domain

import "time"

type SystemSetting struct {
	Key         string    `json:"key" db:"setting_key"`
	Value       string    `json:"value" db:"setting_value"`
	Description string    `json:"description" db:"description"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type AdminLog struct {
	ID        UUID      `json:"id" db:"id"`
	AdminID   *UUID     `json:"adminId,omitempty" db:"admin_id"`
	Action    string    `json:"action" db:"action"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
