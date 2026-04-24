package wallet

import (
	"context"
	"strings"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	txRepo     ports.TransactionRepository
	walletRepo ports.WalletRepository
	systemRepo ports.SystemSettingsRepository
}

func NewUseCase(wr ports.WalletRepository, tr ports.TransactionRepository, systemRepo ports.SystemSettingsRepository) *UseCase {
	return &UseCase{
		txRepo:     tr,
		walletRepo: wr,
		systemRepo: systemRepo,
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

// GetAutoTopUpEnabled — текущее состояние глобального автотопапа кошелька.
func (uc *UseCase) GetAutoTopUpEnabled(ctx context.Context, userID domain.UUID) (bool, error) {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, wrapper.Wrap(err)
	}

	return wallet.AutoTopUpEnabled, nil
}

// TopUpWallet — пополнение кошелька (канал СБП; при отключённом флаге — ошибка).
func (uc *UseCase) TopUpWallet(ctx context.Context, userID domain.UUID, amount domain.Numeric) error {
	if uc.systemRepo != nil {
		enabled, err := sbpTopupEnabled(ctx, uc.systemRepo)
		if err != nil {
			return wrapper.Wrap(err)
		}
		if !enabled {
			return domain.NewSBPTopUpDisabled()
		}
	}

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

func sbpTopupEnabled(ctx context.Context, repo ports.SystemSettingsRepository) (bool, error) {
	list, err := repo.ListAll(ctx)
	if err != nil {
		return true, wrapper.Wrap(err)
	}

	for _, row := range list {
		if row == nil {
			continue
		}
		if row.Key == "sbp_topup_enabled" {
			if row.BoolValue != nil {
				return *row.BoolValue, nil
			}
			// Fallback: text values
			v := strings.TrimSpace(strings.ToLower(row.Value))
			if v == "0" || v == "false" || v == "off" {
				return false, nil
			}

			return true, nil
		}
	}

	return true, nil
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
