package auth

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=contracts.go -destination=./mocks/contracts_mock.go -package=mocks

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

// AuthRegisterLogin — регистрация и вход (для HTTP-слоя и gomock-тестов хендлера).
type AuthRegisterLogin interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
}

// WalletBalanceProvider — баланс в ответах auth-хендлера.
type WalletBalanceProvider interface {
	GetBalance(ctx context.Context, userID domain.UUID) (domain.Numeric, error)
}

// UserByIDReader — пользователь по id (refresh token).
type UserByIDReader interface {
	GetByID(ctx context.Context, id domain.UUID) (*domain.User, error)
}
