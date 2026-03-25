package service

import (
	"time"
)

// CardProvider - интерфейс для работы с провайдерами карт (Wallester, IDBank, Evoca и т.д.)
// Позволяет легко переключаться между банками-эмитентами
type CardProvider interface {
	// GetCardDetails получает детали карты (номер, CVV, срок действия)
	GetCardDetails(cardID int) (*CardDetails, error)

	// IssueCard выпускает новую карту через провайдера
	IssueCard(request IssueCardRequest) (*IssuedCard, error)

	// TopUpCard пополняет баланс карты
	TopUpCard(cardID int, amount float64, currency string) error

	// FreezeCard замораживает карту
	FreezeCard(cardID int) error

	// UnfreezeCard размораживает карту
	UnfreezeCard(cardID int) error

	// GetProviderName возвращает имя провайдера (для логирования)
	GetProviderName() string
}

// CardDetails - детали карты от провайдера
type CardDetails struct {
	CardNumber     string    `json:"card_number"`
	CVV            string    `json:"cvv"`
	ExpiryDate     string    `json:"expiry_date"` // MM/YY
	HolderName     string    `json:"holder_name"`
	Balance        float64   `json:"balance"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	Last4          string    `json:"last_4"`
	BIN            string    `json:"bin"`
	ProviderCardID string    `json:"provider_card_id"` // ID карты в системе провайдера
	CreatedAt      time.Time `json:"created_at"`
}

// IssueCardRequest - запрос на выпуск карты
type IssueCardRequest struct {
	UserID       int     `json:"user_id"`
	CardType     string  `json:"card_type"`     // subscriptions, travel, premium, universal
	Currency     string  `json:"currency"`      // USD, EUR
	DailyLimit   float64 `json:"daily_limit"`
	MonthlyLimit float64 `json:"monthly_limit"`
	HolderName   string  `json:"holder_name"`
	Category     string  `json:"category"` // arbitrage, personal, etc.
}

// IssuedCard - результат выпуска карты
type IssuedCard struct {
	ProviderCardID string    `json:"provider_card_id"`
	CardNumber     string    `json:"card_number"`
	CVV            string    `json:"cvv"`
	ExpiryDate     string    `json:"expiry_date"`
	Last4          string    `json:"last_4"`
	BIN            string    `json:"bin"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// ProviderError - ошибка от провайдера
type ProviderError struct {
	Provider string
	Code     string
	Message  string
}

func (e *ProviderError) Error() string {
	return e.Provider + " error [" + e.Code + "]: " + e.Message
}
