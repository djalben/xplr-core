package ports

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type WalletRepository interface {
	GetByUserID(ctx context.Context, userID domain.UUID) (*domain.Wallet, error)
	Update(ctx context.Context, wallet *domain.Wallet) error
}
