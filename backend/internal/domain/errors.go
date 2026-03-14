package domain

import (
	"errors"

	"gitlab.com/libs-artifex/wrapper/v2"
)

// Общие ошибки домена (базовые sentinel errors).
var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
)

// NewInvalidInput — теперь с константным форматом.
func NewInvalidInput(msg string) error {
	return wrapper.Wrapf(ErrInvalidInput, "%s", msg)
}

// NewInsufficientFunds — без параметров, всё ок.
func NewInsufficientFunds() error {
	return wrapper.Wrap(ErrInsufficientFunds)
}

// NewNotFound. Пример дополнительных (если понадобится).
func NewNotFound(msg string) error {
	return wrapper.Wrapf(ErrNotFound, "%s", msg)
}
