package card

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	cardRepo       ports.CardRepository
	walletRepo     ports.WalletRepository
	txRepo         ports.TransactionRepository
	gradeRepo      ports.GradeRepository
	commissionRepo ports.CommissionConfigRepository
}

func NewUseCase(
	cr ports.CardRepository,
	wr ports.WalletRepository,
	tr ports.TransactionRepository,
	gr ports.GradeRepository,
	commissionRepo ports.CommissionConfigRepository,
) *UseCase {
	return &UseCase{
		cardRepo:       cr,
		walletRepo:     wr,
		txRepo:         tr,
		gradeRepo:      gr,
		commissionRepo: commissionRepo,
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

	issueFee, err := uc.cardIssueFee(ctx)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	wallet, err := uc.chargeCardIssueFee(ctx, userID, issueFee)
	if err != nil {
		return nil, err
	}

	card, err := domain.NewCard(userID, cardType, currency, "TEMP_PROVIDER_ID", nickname)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	err = uc.cardRepo.Save(ctx, card)
	if err != nil {
		saveErr := wrapper.Wrap(err)

		refundErr := uc.refundCardIssueFee(ctx, wallet, issueFee)
		if refundErr != nil {
			return nil, wrapper.Wrap(errors.Join(saveErr, refundErr))
		}

		return nil, saveErr
	}

	if issueFee.GreaterThan(domain.NewNumeric(0)) {
		tx := domain.NewTransaction(userID, &card.ID, issueFee, domain.NewNumeric(0),
			"CARD_ISSUE", "COMPLETED", "Выпуск виртуальной карты")

		err = uc.txRepo.Save(ctx, tx)
		if err != nil {
			return nil, wrapper.Wrap(err)
		}
	}

	return card, nil
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

	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
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
		updateCardErr := wrapper.Wrap(err)

		refundErr := wallet.TopUp(amount)
		if refundErr == nil {
			refundErr = uc.walletRepo.Update(ctx, wallet)
		}
		if refundErr != nil {
			return wrapper.Wrap(errors.Join(updateCardErr, refundErr))
		}

		return updateCardErr
	}

	tx := domain.NewTransaction(userID, &cardID, amount, domain.NewNumeric(0),
		"TOPUP_CARD", "COMPLETED", "Ручное пополнение карты")

	return uc.txRepo.Save(ctx, tx)
}

// SpendFromCard — списание с баланса карты с проверкой дневного лимита (на карту) и месячного (по типу карты).
func (uc *UseCase) SpendFromCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, amount domain.Numeric) error {
	return uc.SpendFromCardWithDetails(ctx, userID, cardID, amount, "Списание с карты")
}

// SpendFromCardWithDetails — списание с баланса карты с проверкой лимитов + запись details (например, merchant/service).
func (uc *UseCase) SpendFromCardWithDetails(ctx context.Context, userID domain.UUID, cardID domain.UUID, amount domain.Numeric, details string) error {
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

	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	tx := domain.NewTransaction(userID, &cardID, amount, domain.NewNumeric(0),
		domain.TransactionTypeCardSpend, "COMPLETED", details)

	return uc.txRepo.Save(ctx, tx)
}

// CloseCard — закрытие карты + возврат остатка на кошелёк.
func (uc *UseCase) CloseCard(ctx context.Context, userID domain.UUID, cardID domain.UUID) error {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if card.UserID != userID {
		return domain.NewInvalidInput("card not found")
	}
	if card.CardStatus == domain.CardStatusClosed {
		return nil
	}

	refund := card.Balance
	card.Balance = domain.NewNumeric(0)
	card.CardStatus = domain.CardStatusClosed

	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if !refund.GreaterThan(domain.NewNumeric(0)) {
		return nil
	}

	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = wallet.TopUp(refund)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = uc.walletRepo.Update(ctx, wallet)
	if err != nil {
		return wrapper.Wrap(err)
	}

	tx := domain.NewTransaction(userID, &cardID, refund, domain.NewNumeric(0),
		domain.TransactionTypeCardRefund, "COMPLETED", "Возврат остатка при закрытии карты")

	return uc.txRepo.Save(ctx, tx)
}

// AutoTopUpCard — автопополнение карты с кошелька (при включённом auto top-up).
func (uc *UseCase) AutoTopUpCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, neededAmount domain.Numeric) error {
	if neededAmount.LessThanOrEqual(domain.NewNumeric(0)) {
		return domain.NewInvalidInput("amount must be positive")
	}

	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}
	if !wallet.AutoTopUpEnabled {
		return domain.NewInvalidInput("auto top-up is disabled on wallet")
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

	err = wallet.Withdraw(neededAmount)
	if err != nil {
		return wrapper.Wrap(err)
	}

	card.Balance = card.Balance.Add(neededAmount)

	err = uc.walletRepo.Update(ctx, wallet)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		updateCardErr := wrapper.Wrap(err)

		refundErr := wallet.TopUp(neededAmount)
		if refundErr == nil {
			refundErr = uc.walletRepo.Update(ctx, wallet)
		}
		if refundErr != nil {
			return wrapper.Wrap(errors.Join(updateCardErr, refundErr))
		}

		return updateCardErr
	}

	tx := domain.NewTransaction(userID, &cardID, neededAmount, domain.NewNumeric(0),
		"AUTO_TOPUP", "COMPLETED", "Автоматическое пополнение карты с кошелька")

	return uc.txRepo.Save(ctx, tx)
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

	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// UnblockCard — разблокирует карту (только владелец).
func (uc *UseCase) UnblockCard(ctx context.Context, userID domain.UUID, cardID domain.UUID) error {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if card.UserID != userID {
		return domain.NewInvalidInput("card not found")
	}

	err = rejectUnblockClosedCard(card)
	if err != nil {
		return err
	}

	card.FailedAuthCount = 0
	card.CardStatus = domain.CardStatusActive

	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// UnblockCardAdmin — разблокирует карту без проверки владельца (только админ-эндпоинт).
func (uc *UseCase) UnblockCardAdmin(ctx context.Context, cardID domain.UUID) error {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = rejectUnblockClosedCard(card)
	if err != nil {
		return err
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

	err = uc.cardRepo.Update(ctx, card)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// UpdateStatus — меняет статус карты: CLOSED — закрытие; ACTIVE — только разблокировка BLOCKED (не реактивация CLOSED).
func (uc *UseCase) UpdateStatus(ctx context.Context, userID domain.UUID, cardID domain.UUID, status string) error {
	if status == domain.CardStatusClosed {
		return uc.CloseCard(ctx, userID, cardID)
	}

	if status == domain.CardStatusActive {
		return uc.UnblockCard(ctx, userID, cardID)
	}

	return domain.NewInvalidInput("invalid status")
}

func rejectUnblockClosedCard(card *domain.Card) error {
	if card.CardStatus == domain.CardStatusClosed {
		return domain.NewInvalidInput("card is closed")
	}

	return nil
}

func (uc *UseCase) chargeCardIssueFee(ctx context.Context, userID domain.UUID, fee domain.Numeric) (*domain.Wallet, error) {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	if !fee.GreaterThan(domain.NewNumeric(0)) {
		return wallet, nil
	}

	if wallet.Balance.LessThan(fee) {
		return nil, domain.NewInsufficientFunds()
	}

	err = wallet.Withdraw(fee)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	err = uc.walletRepo.Update(ctx, wallet)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return wallet, nil
}

func (uc *UseCase) refundCardIssueFee(ctx context.Context, wallet *domain.Wallet, fee domain.Numeric) error {
	if !fee.GreaterThan(domain.NewNumeric(0)) {
		return nil
	}

	err := wallet.TopUp(fee)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = uc.walletRepo.Update(ctx, wallet)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (uc *UseCase) cardIssueFee(ctx context.Context) (domain.Numeric, error) {
	defaultFee := domain.NewNumeric(2)

	if uc.commissionRepo == nil {
		return defaultFee, nil
	}

	cfg, err := uc.commissionRepo.GetByKey(ctx, domain.CardIssueFee)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultFee, nil
		}

		return domain.NewNumeric(0), wrapper.Wrap(err)
	}

	if cfg.Value.LessThanOrEqual(domain.NewNumeric(0)) {
		return defaultFee, nil
	}

	return cfg.Value, nil
}
