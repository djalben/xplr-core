package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

// AuthUseCase — интерфейс для регистрации и входа.
type AuthUseCase interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
}
