package domain

import "time"

type CardType string

const (
	CardTypeSubscriptions CardType = "subscriptions"
	CardTypeTravel        CardType = "travel"
	CardTypePremium       CardType = "premium"
)

type CardCurrency string

const (
	CardCurrencyUSD CardCurrency = "USD"
	CardCurrencyEUR CardCurrency = "EUR"
)

// Статусы карты (card_status в БД).
const (
	CardStatusActive  = "ACTIVE"
	CardStatusClosed  = "CLOSED"
	CardStatusBlocked = "BLOCKED"
)

// AntifraudMaxFailedAuthAttempts — после стольких неудачных авторизаций карта блокируется.
const AntifraudMaxFailedAuthAttempts int64 = 3

type Card struct {
	ID              UUID         `json:"id" db:"id"`
	UserID          UUID         `json:"userId" db:"user_id"`
	ProviderCardID  string       `json:"providerCardId" db:"provider_card_id"`
	Bin             string       `json:"bin" db:"bin"`
	Last4Digits     string       `json:"last4Digits" db:"last_4_digits"`
	CardStatus      string       `json:"cardStatus" db:"card_status"`
	Nickname        string       `json:"nickname" db:"nickname"`
	DailySpendLimit Numeric      `json:"dailySpendLimit" db:"daily_spend_limit"`
	FailedAuthCount int64        `json:"failedAuthCount" db:"failed_auth_count"`
	CardType        CardType     `json:"cardType" db:"card_type"`
	Currency        CardCurrency `json:"currency" db:"currency"`
	Balance         Numeric      `json:"balance" db:"balance"`
	ExpiryDate      *time.Time   `json:"expiryDate" db:"expiry_date"`
	CreatedAt       time.Time    `json:"createdAt" db:"created_at"`
}

func NewCard(userID UUID, cardType CardType, currency CardCurrency, providerCardID string, nickname string) (*Card, error) {
	if !isValidCardType(cardType) {
		return nil, NewInvalidInput("invalid card_type: must be subscriptions, travel or premium")
	}
	if !isValidCurrency(currency) {
		return nil, NewInvalidInput("invalid currency: must be USD or EUR")
	}

	return &Card{
		ID:              NewUUID(),
		UserID:          userID,
		ProviderCardID:  providerCardID,
		Bin:             "424242",
		Last4Digits:     "0000",
		CardStatus:      CardStatusActive,
		Nickname:        nickname,
		CardType:        cardType,
		Currency:        currency,
		DailySpendLimit: NewNumeric(1000),
		Balance:         NewNumeric(0),
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func isValidCardType(t CardType) bool {
	return t == CardTypeSubscriptions || t == CardTypeTravel || t == CardTypePremium
}

func isValidCurrency(c CardCurrency) bool {
	return c == CardCurrencyUSD || c == CardCurrencyEUR
}
