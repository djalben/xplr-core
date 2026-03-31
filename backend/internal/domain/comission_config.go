package domain

import "time"

type CommissionConfig struct {
	ID          UUID      `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	Value       Numeric   `json:"value" db:"value"`
	Description string    `json:"description" db:"description"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

const (
	FeeStandard     = "fee_standard"
	FeeGold         = "fee_gold"
	ReferralPercent = "referral_percent"
	CardIssueFee    = "card_issue_fee"
)
