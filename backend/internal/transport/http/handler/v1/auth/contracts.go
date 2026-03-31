package auth

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=contracts.go -destination=./mocks/contracts_mock.go -package=mocks

import (
	"context"

	authapp "github.com/djalben/xplr-core/backend/internal/application/auth"
	"github.com/djalben/xplr-core/backend/internal/domain"
)

// AuthFlow — регистрация, вход, MFA, email, сброс пароля.
type AuthFlow interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*authapp.LoginResult, error)
	CompleteMFALogin(ctx context.Context, mfaToken, totpCode string) (*domain.User, error)
	VerifyEmail(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// WalletBalanceProvider — баланс в ответах auth-хендлера.
type WalletBalanceProvider interface {
	GetBalance(ctx context.Context, userID domain.UUID) (domain.Numeric, error)
}

// UserByIDReader — пользователь по id (refresh token).
type UserByIDReader interface {
	GetByID(ctx context.Context, id domain.UUID) (*domain.User, error)
}
