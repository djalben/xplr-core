package service

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"
)

// MockProvider - временная заглушка, работающая на данных из БД XPLR
// Используется до интеграции с реальным армянским банком
type MockProvider struct {
	db *sql.DB
}

// NewMockProvider создает новый mock-провайдер
func NewMockProvider(db *sql.DB) *MockProvider {
	return &MockProvider{db: db}
}

// GetProviderName возвращает имя провайдера
func (m *MockProvider) GetProviderName() string {
	return "MockProvider"
}

// GetCardDetails получает детали карты из БД XPLR
func (m *MockProvider) GetCardDetails(cardID int) (*CardDetails, error) {
	var details CardDetails
	var cardNumber, cvv, expiryDate, holderName, currency, status, last4 sql.NullString
	var balance sql.NullFloat64
	var createdAt sql.NullTime

	query := `
		SELECT 
			COALESCE(card_number, '') as card_number,
			COALESCE(cvv, '') as cvv,
			COALESCE(expiry_date::text, '') as expiry_date,
			COALESCE(holder_name, 'XPLR CARDHOLDER') as holder_name,
			COALESCE(card_balance, 0) as balance,
			COALESCE(currency, 'USD') as currency,
			COALESCE(card_status, 'active') as status,
			COALESCE(last_4, '') as last_4,
			created_at
		FROM cards
		WHERE id = $1
	`

	err := m.db.QueryRow(query, cardID).Scan(
		&cardNumber, &cvv, &expiryDate, &holderName,
		&balance, &currency, &status, &last4, &createdAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ProviderError{
				Provider: "MockProvider",
				Code:     "CARD_NOT_FOUND",
				Message:  fmt.Sprintf("Card %d not found", cardID),
			}
		}
		return nil, err
	}

	details.CardNumber = cardNumber.String
	details.CVV = cvv.String
	details.ExpiryDate = expiryDate.String
	details.HolderName = holderName.String
	details.Balance = balance.Float64
	details.Currency = currency.String
	details.Status = status.String
	details.Last4 = last4.String
	details.ProviderCardID = fmt.Sprintf("MOCK-%d", cardID)
	
	if createdAt.Valid {
		details.CreatedAt = createdAt.Time
	}

	// Если номер карты пустой, генерируем mock-номер
	if details.CardNumber == "" {
		details.CardNumber = m.generateMockCardNumber()
		details.Last4 = details.CardNumber[len(details.CardNumber)-4:]
	}

	// Если CVV пустой, генерируем mock-CVV
	if details.CVV == "" {
		details.CVV = m.generateMockCVV()
	}

	// Если срок действия пустой, генерируем на 3 года вперед
	if details.ExpiryDate == "" {
		futureDate := time.Now().AddDate(3, 0, 0)
		details.ExpiryDate = futureDate.Format("01/06")[0:5] // MM/YY
	}

	// BIN для mock-карт
	details.BIN = details.CardNumber[0:6]

	log.Printf("[MOCK-PROVIDER] GetCardDetails: card_id=%d, last4=%s, balance=%.2f %s", 
		cardID, details.Last4, details.Balance, details.Currency)

	return &details, nil
}

// IssueCard выпускает mock-карту (сохраняет в БД XPLR)
func (m *MockProvider) IssueCard(request IssueCardRequest) (*IssuedCard, error) {
	cardNumber := m.generateMockCardNumber()
	cvv := m.generateMockCVV()
	futureDate := time.Now().AddDate(3, 0, 0)
	expiryDate := futureDate.Format("01/06")[0:5] // MM/YY
	last4 := cardNumber[len(cardNumber)-4:]
	bin := cardNumber[0:6]

	issued := &IssuedCard{
		ProviderCardID: fmt.Sprintf("MOCK-%d-%d", request.UserID, time.Now().Unix()),
		CardNumber:     cardNumber,
		CVV:            cvv,
		ExpiryDate:     expiryDate,
		Last4:          last4,
		BIN:            bin,
		Status:         "active",
		CreatedAt:      time.Now(),
	}

	log.Printf("[MOCK-PROVIDER] IssueCard: user_id=%d, card_type=%s, currency=%s, last4=%s", 
		request.UserID, request.CardType, request.Currency, last4)

	return issued, nil
}

// TopUpCard пополняет баланс карты в БД
func (m *MockProvider) TopUpCard(cardID int, amount float64, currency string) error {
	_, err := m.db.Exec(`
		UPDATE cards 
		SET card_balance = card_balance + $1,
		    updated_at = NOW()
		WHERE id = $2
	`, amount, cardID)

	if err != nil {
		return &ProviderError{
			Provider: "MockProvider",
			Code:     "TOPUP_FAILED",
			Message:  fmt.Sprintf("Failed to top up card %d: %v", cardID, err),
		}
	}

	log.Printf("[MOCK-PROVIDER] TopUpCard: card_id=%d, amount=%.2f %s", cardID, amount, currency)
	return nil
}

// FreezeCard замораживает карту
func (m *MockProvider) FreezeCard(cardID int) error {
	_, err := m.db.Exec(`
		UPDATE cards 
		SET card_status = 'frozen',
		    updated_at = NOW()
		WHERE id = $1
	`, cardID)

	if err != nil {
		return &ProviderError{
			Provider: "MockProvider",
			Code:     "FREEZE_FAILED",
			Message:  fmt.Sprintf("Failed to freeze card %d: %v", cardID, err),
		}
	}

	log.Printf("[MOCK-PROVIDER] FreezeCard: card_id=%d", cardID)
	return nil
}

// UnfreezeCard размораживает карту
func (m *MockProvider) UnfreezeCard(cardID int) error {
	_, err := m.db.Exec(`
		UPDATE cards 
		SET card_status = 'active',
		    updated_at = NOW()
		WHERE id = $1
	`, cardID)

	if err != nil {
		return &ProviderError{
			Provider: "MockProvider",
			Code:     "UNFREEZE_FAILED",
			Message:  fmt.Sprintf("Failed to unfreeze card %d: %v", cardID, err),
		}
	}

	log.Printf("[MOCK-PROVIDER] UnfreezeCard: card_id=%d", cardID)
	return nil
}

// generateMockCardNumber генерирует тестовый номер карты (16 цифр)
func (m *MockProvider) generateMockCardNumber() string {
	// Используем BIN 555555 (Mastercard test range)
	bin := "555555"
	
	// Генерируем 9 случайных цифр
	middle := ""
	for i := 0; i < 9; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		middle += fmt.Sprintf("%d", n.Int64())
	}
	
	// Последняя цифра - контрольная (упрощенно, просто случайная)
	checkDigit, _ := rand.Int(rand.Reader, big.NewInt(10))
	
	return bin + middle + fmt.Sprintf("%d", checkDigit.Int64())
}

// generateMockCVV генерирует тестовый CVV (3 цифры)
func (m *MockProvider) generateMockCVV() string {
	cvv := ""
	for i := 0; i < 3; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		cvv += fmt.Sprintf("%d", n.Int64())
	}
	return cvv
}
