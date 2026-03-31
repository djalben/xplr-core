package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id domain.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByEmailVerifyTokenHash(ctx context.Context, tokenHash string) (*domain.User, error)
	GetByPasswordResetTokenHash(ctx context.Context, tokenHash string) (*domain.User, error)
	GetByTelegramChatID(ctx context.Context, chatID int64) (*domain.User, error)
	Save(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
}
