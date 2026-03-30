package wallet

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	walletRepo ports.WalletRepository
	txRepo     ports.TransactionRepository
}

func NewUseCase(wr ports.WalletRepository, tr ports.TransactionRepository) *UseCase {
	return &UseCase{
		walletRepo: wr,
		txRepo:     tr,
	}
}

// GetBalance — получение баланса кошелька.
func (uc *UseCase) GetBalance(ctx context.Context, userID domain.UUID) (domain.Numeric, error) {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return domain.NewNumeric(0), wrapper.Wrap(err)
	}

	return wallet.Balance, nil
}

// TopUpWallet — пополнение кошелька.
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

	tx := domain.NewTransaction(
		userID,
		nil,
		amount,
		domain.NewNumeric(0),
		"TOPUP_WALLET",
		"COMPLETED",
		"Пополнение кошелька",
	)

	return uc.txRepo.Save(ctx, tx)
}

// ToggleAutoTopUp — включить/выключить глобальный автотопап.
func (uc *UseCase) ToggleAutoTopUp(ctx context.Context, userID domain.UUID, enabled bool) error {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	wallet.AutoTopUpEnabled = enabled

	err = uc.walletRepo.Update(ctx, wallet)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// AutoTopUpCard — автоматическое пополнение карты с кошелька (вызывается при недостатке средств).
func (uc *UseCase) AutoTopUpCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, neededAmount domain.Numeric) error {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if !wallet.AutoTopUpEnabled {
		return domain.NewInvalidInput("auto top-up is disabled on wallet")
	}

	err = wallet.Withdraw(neededAmount)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = uc.walletRepo.Update(ctx, wallet)
	if err != nil {
		return wrapper.Wrap(err)
	}

	tx := domain.NewTransaction(
		userID,
		&cardID,
		neededAmount,
		domain.NewNumeric(0),
		"AUTO_TOPUP",
		"COMPLETED",
		"Автоматическое пополнение карты с кошелька",
	)

	return uc.txRepo.Save(ctx, tx)
}
