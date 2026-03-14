package wallet

import (
	"context"

	"github.com/djalben/xplr-core/internal/domain"
	"github.com/djalben/xplr-core/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	walletRepo ports.WalletRepository
	txRepo     ports.TransactionRepository
}

func NewUseCase(wr ports.WalletRepository, tr ports.TransactionRepository) *UseCase {
	return &UseCase{walletRepo: wr, txRepo: tr}
}

func (uc *UseCase) TopUpWallet(ctx context.Context, userID domain.UUID, amount domain.Numeric) error {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = wallet.TopUp(amount)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = uc.walletRepo.Update(ctx, wallet)
	if err != nil {
		return wrapper.Wrap(err)
	}

	tx := domain.NewTransaction(userID, nil, amount, domain.NewNumeric(0), "TOPUP_WALLET", "COMPLETED", "Пополнение по СБП")

	return wrapper.Wrap(uc.txRepo.Save(ctx, tx))
}

func (uc *UseCase) GetBalance(ctx context.Context, userID domain.UUID) (domain.Numeric, error) {
	w, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return domain.NewNumeric(0), wrapper.Wrap(err)
	}

	return w.Balance, nil
}
