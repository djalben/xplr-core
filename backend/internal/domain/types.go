package domain

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// UUID — наш внутренний тип ID (alias, чтобы везде можно было использовать как uuid.UUID).
type UUID = uuid.UUID

// NewUUID — удобная фабрика.
func NewUUID() UUID {
	return uuid.New()
}

// Numeric — деньги без плавающей точки (самый правильный способ в финтехе).
type Numeric = decimal.Decimal

// NewNumeric — создаём из float (или можно из строки, если захочешь).
func NewNumeric(value float64) Numeric {
	return decimal.NewFromFloat(value)
}
