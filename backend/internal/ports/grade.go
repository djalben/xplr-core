package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type GradeRepository interface {
	GetByUserID(ctx context.Context, userID domain.UUID) (*domain.UserGrade, error)
	Update(ctx context.Context, grade *domain.UserGrade) error
}
