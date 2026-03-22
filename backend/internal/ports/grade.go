package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type GradeRepository interface {
	GetByUserID(ctx context.Context, userID domain.UUID) (*domain.UserGrade, error)
	EnsureGrade(ctx context.Context, userID domain.UUID) error
	Update(ctx context.Context, grade *domain.UserGrade) error
}
