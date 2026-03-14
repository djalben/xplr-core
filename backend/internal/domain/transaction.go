package domain

import (
	"time"
)

type Transaction struct {
	ID              UUID      `json:"id" db:"id"`
	UserID          UUID      `json:"userId" db:"user_id"`
	CardID          *UUID     `json:"cardId,omitempty" db:"card_id"`
	Amount          Numeric   `json:"amount" db:"amount"`
	Fee             Numeric   `json:"fee" db:"fee"`
	TransactionType string    `json:"transactionType" db:"transaction_type"`
	Status          string    `json:"status" db:"status"`
	Details         string    `json:"details" db:"details"`
	ProviderTxID    string    `json:"providerTxId,omitempty" db:"provider_tx_id"`
	ExecutedAt      time.Time `json:"executedAt" db:"executed_at"`
}

func NewTransaction(
	userID UUID,
	cardID *UUID,
	amount, fee Numeric,
	txType, status, details string,
) *Transaction {
	return &Transaction{
		ID:              NewUUID(),
		UserID:          userID,
		CardID:          cardID,
		Amount:          amount,
		Fee:             fee,
		TransactionType: txType,
		Status:          status,
		Details:         details,
		ExecutedAt:      time.Now().UTC(),
	}
}
