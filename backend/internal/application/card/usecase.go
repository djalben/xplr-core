package card

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/application/wallet"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	cardRepo   ports.CardRepository
	walletRepo ports.WalletRepository
	txRepo     ports.TransactionRepository
	walletUC   *wallet.UseCase
}

func NewUseCase(
	cr ports.CardRepository,
	wr ports.WalletRepository,
	tr ports.TransactionRepository,
	walletUC *wallet.UseCase,
) *UseCase {
	return &UseCase{
		cardRepo:   cr,
		walletRepo: wr,
		txRepo:     tr,
		walletUC:   walletUC,
	}
}

// BuyCard — выпуск новой карты.
func (uc *UseCase) BuyCard(ctx context.Context, userID domain.UUID, cardType domain.CardType) (*domain.Card, error) {
	card, err := domain.NewCard(userID, cardType, "TEMP_PROVIDER_ID")
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	err = uc.cardRepo.Save(ctx, card)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	tx := domain.NewTransaction(userID, &card.ID, domain.NewNumeric(2.00), domain.NewNumeric(0),
		"CARD_ISSUE", "COMPLETED", "Выпуск виртуальной карты")
	return card, uc.txRepo.Save(ctx, tx)
}

// GetByID — получение карты по ID.
func (uc *UseCase) GetByID(ctx context.Context, cardID domain.UUID) (*domain.Card, error) {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return card, nil
}

// TopUpCard — ручное пополнение карты.
func (uc *UseCase) TopUpCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, amount domain.Numeric) error {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = wallet.Withdraw(amount)
	if err != nil {
		return wrapper.Wrap(err)
	}

	card.Balance = card.Balance.Add(amount)

	err = uc.walletRepo.Update(ctx, wallet)
	if err != nil {
		return wrapper.Wrap(err)
	}
	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	tx := domain.NewTransaction(userID, &cardID, amount, domain.NewNumeric(0),
		"TOPUP_CARD", "COMPLETED", "Ручное пополнение карты")
	return uc.txRepo.Save(ctx, tx)
}

// CloseCard — закрытие карты + возврат остатка на кошелёк.
func (uc *UseCase) CloseCard(ctx context.Context, userID domain.UUID, cardID domain.UUID) error {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if card.Balance.GreaterThan(domain.NewNumeric(0)) {
		wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
		if err != nil {
			return wrapper.Wrap(err)
		}

		err = wallet.TopUp(card.Balance)
		if err != nil {
			return wrapper.Wrap(err)
		}

		err = uc.walletRepo.Update(ctx, wallet)
		if err != nil {
			return wrapper.Wrap(err)
		}
	}

	card.CardStatus = "CLOSED"
	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// AutoTopUpCard — главный метод автотопапа (вызывается, когда карте не хватило денег).
func (uc *UseCase) AutoTopUpCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, neededAmount domain.Numeric) error {
	return uc.walletUC.AutoTopUpCard(ctx, userID, cardID, neededAmount)
}
