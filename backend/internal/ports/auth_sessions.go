package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type AuthSessionsRepository interface {
	Add(ctx context.Context, s *domain.AuthSession) error
	ListByUserID(ctx context.Context, userID domain.UUID, limit int) ([]*domain.AuthSession, error)
	DeleteByUserID(ctx context.Context, userID domain.UUID) error
	DeleteOlderThan(ctx context.Context, userID domain.UUID, keepLast int) error
}
