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
	// ErrCardBlockedAntifraud — карта заблокирована после серии неудачных авторизаций.
	ErrCardBlockedAntifraud = errors.New("card blocked: too many failed authorization attempts")
	// ErrSBPTopUpDisabled — глобально отключено пополнение через СБП.
	ErrSBPTopUpDisabled = errors.New("sbp top-up is disabled")
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

// NewSBPTopUpDisabled — пополнение через СБП недоступно.
func NewSBPTopUpDisabled() error {
	return wrapper.Wrap(ErrSBPTopUpDisabled)
}

// NewCardBlockedAntifraud — карта заблокирована антифродом.
func NewCardBlockedAntifraud() error {
	return wrapper.Wrap(ErrCardBlockedAntifraud)
}

// NewAlreadyExists — сущность уже существует (например, занятый telegram_chat_id).
func NewAlreadyExists(msg string) error {
	return wrapper.Wrapf(ErrAlreadyExists, "%s", msg)
}
