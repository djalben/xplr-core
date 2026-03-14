package domain

import (
	"time"
)

type CardType string

const (
	CardTypeSubscriptions CardType = "subscriptions"
	CardTypeTravel        CardType = "travel"
	CardTypePremium       CardType = "premium"
)

type Card struct {
	ID               UUID       `json:"id"               db:"id"`
	UserID           UUID       `json:"userId"           db:"user_id"`
	ProviderCardID   string     `json:"providerCardId"   db:"provider_card_id"`
	Bin              string     `json:"bin"              db:"bin"`
	Last4Digits      string     `json:"last4Digits"      db:"last_4_digits"`
	CardStatus       string     `json:"cardStatus"       db:"card_status"`
	Nickname         string     `json:"nickname"         db:"nickname"`
	DailySpendLimit  Numeric    `json:"dailySpendLimit"  db:"daily_spend_limit"`
	FailedAuthCount  int64      `json:"failedAuthCount"  db:"failed_auth_count"`
	CardType         CardType   `json:"cardType"         db:"card_type"`
	AutoTopUpEnabled bool       `json:"autoTopUpEnabled" db:"auto_topup_enabled"`
	AutoTopUpBelow   Numeric    `json:"autoTopUpBelow"   db:"auto_topup_below"`
	AutoTopUpAmount  Numeric    `json:"autoTopUpAmount"  db:"auto_topup_amount"`
	Balance          Numeric    `json:"balance"          db:"balance"`
	ExpiryDate       *time.Time `json:"expiryDate"       db:"expiry_date"`
	CreatedAt        time.Time  `json:"createdAt"        db:"created_at"`
}

func NewCard(userID UUID, cardType CardType, providerCardID string) (*Card, error) {
	if !isValidCardType(cardType) {
		return nil, NewInvalidInput("invalid card_type: must be subscriptions, travel or premium")
	}

	return &Card{
		ID:              NewUUID(),
		UserID:          userID,
		ProviderCardID:  providerCardID,
		Bin:             "424242",
		Last4Digits:     "0000",
		CardStatus:      "ACTIVE",
		CardType:        cardType,
		DailySpendLimit: NewNumeric(1000),
		AutoTopUpBelow:  NewNumeric(100),
		AutoTopUpAmount: NewNumeric(500),
		Balance:         NewNumeric(0),
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func isValidCardType(t CardType) bool {
	return t == CardTypeSubscriptions || t == CardTypeTravel || t == CardTypePremium
}
