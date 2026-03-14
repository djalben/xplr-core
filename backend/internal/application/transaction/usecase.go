package transaction

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	txRepo ports.TransactionRepository
}

func NewUseCase(tr ports.TransactionRepository) *UseCase {
	return &UseCase{txRepo: tr}
}

// GetWalletTransactions — история по кошельку.
func (uc *UseCase) GetWalletTransactions(ctx context.Context, userID domain.UUID, from, to time.Time) ([]*domain.Transaction, error) {
	list, err := uc.txRepo.GetWalletTransactions(ctx, userID, from, to)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return list, nil
}

// GetCardTransactions — история по одной карте.
func (uc *UseCase) GetCardTransactions(ctx context.Context, cardID domain.UUID, from, to time.Time) ([]*domain.Transaction, error) {
	list, err := uc.txRepo.GetCardTransactions(ctx, cardID, from, to)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return list, nil
}
