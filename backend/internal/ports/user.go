package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

//nolint:interfacebloat // Repository contains both read-model and write operations; splitting requires broad refactor.
type UserRepository interface {
	GetByID(ctx context.Context, id domain.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByEmailVerifyTokenHash(ctx context.Context, tokenHash string) (*domain.User, error)
	GetByPasswordResetTokenHash(ctx context.Context, tokenHash string) (*domain.User, error)
	GetByTelegramChatID(ctx context.Context, chatID int64) (*domain.User, error)
	Save(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error

	SearchByEmail(ctx context.Context, query string, limit int) ([]*domain.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error)
	CountUsers(ctx context.Context) (int64, error)
	SetUserStatus(ctx context.Context, id domain.UUID, status domain.UserStatus) error
	SetIsAdmin(ctx context.Context, id domain.UUID, isAdmin bool) error
}
