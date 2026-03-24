package wallet

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=contracts.go -destination=./mocks/contracts_mock.go -package=mocks

// WalletUseCase — сценарии HTTP-слоя /wallet (gomock).
type WalletUseCase interface {
	GetBalance(ctx context.Context, userID domain.UUID) (domain.Numeric, error)
	TopUpWallet(ctx context.Context, userID domain.UUID, amount domain.Numeric) error
	ToggleAutoTopUp(ctx context.Context, userID domain.UUID, enabled bool) error
}
