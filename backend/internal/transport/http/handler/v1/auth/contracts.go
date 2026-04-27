package auth

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=contracts.go -destination=./mocks/contracts_mock.go -package=mocks

import (
	"context"
	"time"

	authapp "github.com/djalben/xplr-core/backend/internal/application/auth"
	"github.com/djalben/xplr-core/backend/internal/domain"
)

// AuthFlow — регистрация, вход, MFA, email, сброс пароля.
type AuthFlow interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*authapp.LoginResult, error)
	LoginWithTrustedDevice(ctx context.Context, email, password, trustedDeviceToken string, now time.Time) (*authapp.LoginResult, error)
	CompleteMFALogin(ctx context.Context, mfaToken, totpCode string) (*domain.User, error)
	RememberTrustedDevice(ctx context.Context, userID domain.UUID, userAgent string, ip *string, now time.Time) (rawToken string, expiresAt time.Time, err error)
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
	SetLastLogin(ctx context.Context, id domain.UUID, at time.Time, ip *string, userAgent string) error
}

// RateLimiter — анти-брутфорс для auth хендлера.
type RateLimiter interface {
	Allow(ctx context.Context, key string, now time.Time) (allowed bool, retryAfter time.Duration, err error)
	Fail(ctx context.Context, key string, now time.Time) (retryAfter time.Duration, err error)
	Success(ctx context.Context, key string, now time.Time) error
}
