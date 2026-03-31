package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

// KYCApplicationRepository — заявки KYC.
type KYCApplicationRepository interface {
	Save(ctx context.Context, app *domain.KYCApplication) error
	GetByID(ctx context.Context, id domain.UUID) (*domain.KYCApplication, error)
	ListByStatus(ctx context.Context, status domain.KYCApplicationStatus, limit int) ([]*domain.KYCApplication, error)
	HasPendingForUser(ctx context.Context, userID domain.UUID) (bool, error)
	GetLatestByUserID(ctx context.Context, userID domain.UUID) (*domain.KYCApplication, error)
	Update(ctx context.Context, app *domain.KYCApplication) error
}
