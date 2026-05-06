package domain

import "time"

type CardSubscription struct {
	ID UUID `json:"id" db:"id"`

	UserID UUID `json:"userId" db:"user_id"`
	CardID UUID `json:"cardId" db:"card_id"`

	MerchantName string `json:"merchantName" db:"merchant_name"`
	MerchantKey  string `json:"merchantKey" db:"merchant_key"`

	LastAmount   Numeric `json:"lastAmount" db:"last_amount"`
	LastCurrency string  `json:"lastCurrency" db:"last_currency"`
	ChargeCount  int     `json:"chargeCount" db:"charge_count"`

	FirstSeenAt time.Time  `json:"firstSeenAt" db:"first_seen_at"`
	LastSeenAt  time.Time  `json:"lastSeenAt" db:"last_seen_at"`
	IsBlocked   bool       `json:"isBlocked" db:"is_blocked"`
	BlockedAt   *time.Time `json:"blockedAt,omitempty" db:"blocked_at"`
}
