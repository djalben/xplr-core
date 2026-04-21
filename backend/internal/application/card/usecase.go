package card

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/application/wallet"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	cardRepo   ports.CardRepository
	walletRepo ports.WalletRepository
	txRepo     ports.TransactionRepository
	gradeRepo  ports.GradeRepository
	walletUC   *wallet.UseCase
}

func NewUseCase(
	cr ports.CardRepository,
	wr ports.WalletRepository,
	tr ports.TransactionRepository,
	gr ports.GradeRepository,
	walletUC *wallet.UseCase,
) *UseCase {
	return &UseCase{
		cardRepo:   cr,
		walletRepo: wr,
		txRepo:     tr,
		gradeRepo:  gr,
		walletUC:   walletUC,
	}
}

// BuyCard — выпуск новой карты.
func (uc *UseCase) BuyCard(ctx context.Context, userID domain.UUID, cardType domain.CardType, nickname string, currency domain.CardCurrency) (*domain.Card, error) {
	err := uc.gradeRepo.EnsureGrade(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	ug, err := uc.gradeRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	maxPerType := domain.MaxActiveCardsPerTypeByGrade(ug.Grade)

	list, err := uc.cardRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	activeSameType := 0

	for _, c := range list {
		if c.CardType == cardType && c.CardStatus != domain.CardStatusClosed {
			activeSameType++
		}
	}

	if activeSameType >= maxPerType {
		return nil, domain.NewInvalidInput("maximum number of active cards of this type for your grade has been reached")
	}

	if currency == "" {
		currency = domain.CardCurrencyUSD
	}
	card, err := domain.NewCard(userID, cardType, currency, "TEMP_PROVIDER_ID", nickname)
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

// ListByUserID — список карт пользователя.
func (uc *UseCase) ListByUserID(ctx context.Context, userID domain.UUID) ([]*domain.Card, error) {
	return uc.cardRepo.ListByUserID(ctx, userID)
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

// SpendFromCard — списание с баланса карты с проверкой дневного лимита (на карту) и месячного (по типу карты).
func (uc *UseCase) SpendFromCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, amount domain.Numeric) error {
	if amount.LessThanOrEqual(domain.NewNumeric(0)) {
		return domain.NewInvalidInput("amount must be positive")
	}

	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if card.UserID != userID {
		return domain.NewInvalidInput("card not found")
	}

	if card.CardStatus != domain.CardStatusActive {
		return domain.NewInvalidInput("card is not active")
	}

	now := time.Now().UTC()

	dayStart, dayEnd := domain.DayBoundsUTC(now)

	spentDay, err := uc.txRepo.SumCardSpendByCardID(ctx, cardID, dayStart, dayEnd)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if card.DailySpendLimit.GreaterThan(domain.NewNumeric(0)) {
		if spentDay.Add(amount).GreaterThan(card.DailySpendLimit) {
			return domain.NewInvalidInput("daily spend limit exceeded")
		}
	}

	monthStart, monthEnd := domain.MonthBoundsUTC(now)

	spentMonthByType, err := uc.txRepo.SumCardSpendByUserAndCardType(ctx, userID, card.CardType, monthStart, monthEnd)
	if err != nil {
		return wrapper.Wrap(err)
	}

	monthCap := domain.MonthlySpendLimitByCardType(card.CardType)

	if monthCap.GreaterThan(domain.NewNumeric(0)) {
		if spentMonthByType.Add(amount).GreaterThan(monthCap) {
			return domain.NewInvalidInput("monthly spend limit exceeded for this card type")
		}
	}

	if card.Balance.LessThan(amount) {
		return domain.NewInsufficientFunds()
	}

	card.Balance = card.Balance.Sub(amount)
	card.FailedAuthCount = 0

	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	tx := domain.NewTransaction(userID, &cardID, amount, domain.NewNumeric(0),
		domain.TransactionTypeCardSpend, "COMPLETED", "Списание с карты")

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

	card.CardStatus = domain.CardStatusClosed
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

// BlockCard — блокирует карту.
func (uc *UseCase) BlockCard(ctx context.Context, cardID domain.UUID) error {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	card.CardStatus = domain.CardStatusBlocked

	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// RecordFailedAuthorization — неудачная авторизация по карте (провайдер); 3 попытки → блокировка.
func (uc *UseCase) RecordFailedAuthorization(ctx context.Context, userID, cardID domain.UUID) error {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if card.UserID != userID {
		return domain.NewInvalidInput("card not found")
	}

	if card.CardStatus != domain.CardStatusActive {
		return domain.NewInvalidInput("card is not active")
	}

	card.FailedAuthCount++
	if card.FailedAuthCount >= domain.AntifraudMaxFailedAuthAttempts {
		card.CardStatus = domain.CardStatusBlocked
	}

	return uc.cardRepo.Update(ctx, card)
}

// UnblockCard — разблокирует карту.
func (uc *UseCase) UnblockCard(ctx context.Context, cardID domain.UUID) error {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	card.FailedAuthCount = 0
	card.CardStatus = domain.CardStatusActive

	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// SetSpendingLimit — устанавливает дневной лимит карты.
func (uc *UseCase) SetSpendingLimit(ctx context.Context, userID domain.UUID, cardID domain.UUID, limit domain.Numeric) error {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if card.UserID != userID {
		return domain.NewInvalidInput("card not found")
	}

	card.DailySpendLimit = limit

	return uc.cardRepo.Update(ctx, card)
}

// UpdateStatus — меняет статус карты (ACTIVE/CLOSED).
func (uc *UseCase) UpdateStatus(ctx context.Context, userID domain.UUID, cardID domain.UUID, status string) error {
	if status == domain.CardStatusClosed {
		return uc.CloseCard(ctx, userID, cardID)
	}

	if status == domain.CardStatusActive {
		return uc.UnblockCard(ctx, cardID)
	}

	return domain.NewInvalidInput("invalid status")
}
