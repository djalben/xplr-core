package service

import (
	"database/sql"
)

// ArmeniaProvider - провайдер для работы с армянскими банками (IDBank, Evoca)
// TODO: Реализовать интеграцию с API армянского банка-эмитента
type ArmeniaProvider struct {
	db         *sql.DB
	apiKey     string
	apiBaseURL string
}

// NewArmeniaProvider создает новый провайдер для армянского банка
func NewArmeniaProvider(db *sql.DB, apiKey string, apiBaseURL string) *ArmeniaProvider {
	return &ArmeniaProvider{
		db:         db,
		apiKey:     apiKey,
		apiBaseURL: apiBaseURL,
	}
}

// GetProviderName возвращает имя провайдера
func (a *ArmeniaProvider) GetProviderName() string {
	return "ArmeniaProvider"
}

// GetCardDetails получает детали карты через API армянского банка
// TODO: Реализовать вызов API
func (a *ArmeniaProvider) GetCardDetails(cardID int) (*CardDetails, error) {
	// TODO: Implement Armenia bank API call
	return nil, &ProviderError{
		Provider: "ArmeniaProvider",
		Code:     "NOT_IMPLEMENTED",
		Message:  "Armenia provider not yet implemented",
	}
}

// IssueCard выпускает карту через API армянского банка
// TODO: Реализовать вызов API
func (a *ArmeniaProvider) IssueCard(request IssueCardRequest) (*IssuedCard, error) {
	// TODO: Implement Armenia bank API call
	return nil, &ProviderError{
		Provider: "ArmeniaProvider",
		Code:     "NOT_IMPLEMENTED",
		Message:  "Armenia provider not yet implemented",
	}
}

// TopUpCard пополняет баланс карты через API армянского банка
// TODO: Реализовать вызов API
func (a *ArmeniaProvider) TopUpCard(cardID int, amount float64, currency string) error {
	// TODO: Implement Armenia bank API call
	return &ProviderError{
		Provider: "ArmeniaProvider",
		Code:     "NOT_IMPLEMENTED",
		Message:  "Armenia provider not yet implemented",
	}
}

// FreezeCard замораживает карту через API армянского банка
// TODO: Реализовать вызов API
func (a *ArmeniaProvider) FreezeCard(cardID int) error {
	// TODO: Implement Armenia bank API call
	return &ProviderError{
		Provider: "ArmeniaProvider",
		Code:     "NOT_IMPLEMENTED",
		Message:  "Armenia provider not yet implemented",
	}
}

// UnfreezeCard размораживает карту через API армянского банка
// TODO: Реализовать вызов API
func (a *ArmeniaProvider) UnfreezeCard(cardID int) error {
	// TODO: Implement Armenia bank API call
	return &ProviderError{
		Provider: "ArmeniaProvider",
		Code:     "NOT_IMPLEMENTED",
		Message:  "Armenia provider not yet implemented",
	}
}
