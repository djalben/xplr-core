package card

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=contracts.go -destination=./mocks/contracts_mock.go -package=mocks

// CardUseCase — сценарии HTTP-слоя /card (gomock).
type CardUseCase interface {
	BuyCard(ctx context.Context, userID domain.UUID, cardType domain.CardType, nickname string) (*domain.Card, error)
	TopUpCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, amount domain.Numeric) error
	SpendFromCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, amount domain.Numeric) error
}
