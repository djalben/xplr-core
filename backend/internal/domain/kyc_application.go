package domain

import "time"

// KYCApplicationStatus — статус заявки KYC.
type KYCApplicationStatus string

const (
	KYCApplicationPending  KYCApplicationStatus = "PENDING"
	KYCApplicationApproved KYCApplicationStatus = "APPROVED"
	KYCApplicationRejected KYCApplicationStatus = "REJECTED"
)

// KYCApplication — заявка пользователя на прохождение KYC.
type KYCApplication struct {
	ID           UUID                 `json:"id" db:"id"`
	UserID       UUID                 `json:"userId" db:"user_id"`
	Status       KYCApplicationStatus `json:"status" db:"status"`
	PayloadJSON  string               `json:"payloadJson" db:"payload_json"`
	AdminComment *string              `json:"adminComment,omitempty" db:"admin_comment"`
	ReviewedBy   *UUID                `json:"reviewedBy,omitempty" db:"reviewed_by"`
	ReviewedAt   *time.Time           `json:"reviewedAt,omitempty" db:"reviewed_at"`
	CreatedAt    time.Time            `json:"createdAt" db:"created_at"`
}

func NewKYCApplication(userID UUID, payloadJSON string) *KYCApplication {
	return &KYCApplication{
		ID:          NewUUID(),
		UserID:      userID,
		Status:      KYCApplicationPending,
		PayloadJSON: payloadJSON,
		CreatedAt:   time.Now().UTC(),
	}
}
