package subscription

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/shopspring/decimal"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type AuthorizationDecision string

const (
	AuthorizationDecisionApprove AuthorizationDecision = "APPROVE"
	AuthorizationDecisionDecline AuthorizationDecision = "DECLINE"
)

type AuthorizationResult struct {
	Decision AuthorizationDecision `json:"decision"`
	Reason   string                `json:"reason,omitempty"`
}

type AuthorizationEvent struct {
	ProviderCardID string `json:"providerCardId"`
	ProviderTxID   string `json:"providerTxId"`
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	MerchantName   string `json:"merchantName"`
	ExecutedAt     string `json:"executedAt,omitempty"` // RFC3339, optional
}

type UseCase struct {
	cardRepo ports.CardRepository
	subRepo  ports.CardSubscriptionRepository
	cardUC   *card.UseCase
}

func NewUseCase(
	cardRepo ports.CardRepository,
	subRepo ports.CardSubscriptionRepository,
	cardUC *card.UseCase,
) *UseCase {
	return &UseCase{
		cardRepo: cardRepo,
		subRepo:  subRepo,
		cardUC:   cardUC,
	}
}

func normalizeMerchantKey(merchantName string) string {
	return strings.ToLower(strings.TrimSpace(merchantName))
}

func (uc *UseCase) List(ctx context.Context, userID domain.UUID) ([]*domain.CardSubscription, error) {
	list, err := uc.subRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return list, nil
}

func (uc *UseCase) SetBlocked(ctx context.Context, userID domain.UUID, subscriptionID domain.UUID, isBlocked bool) error {
	err := uc.subRepo.SetBlocked(ctx, userID, subscriptionID, isBlocked)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (uc *UseCase) SetBlockedByCard(ctx context.Context, userID domain.UUID, cardID domain.UUID, isBlocked bool) error {
	err := uc.subRepo.SetBlockedByCardID(ctx, userID, cardID, isBlocked)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (uc *UseCase) HandleAuthorization(ctx context.Context, event AuthorizationEvent) (AuthorizationResult, error) {
	if strings.TrimSpace(event.ProviderCardID) == "" {
		return AuthorizationResult{}, domain.NewInvalidInput("providerCardId is required")
	}
	if strings.TrimSpace(event.MerchantName) == "" {
		return AuthorizationResult{}, domain.NewInvalidInput("merchantName is required")
	}

	card, err := uc.cardRepo.GetByProviderCardID(ctx, event.ProviderCardID)
	if err != nil {
		return AuthorizationResult{}, wrapper.Wrap(err)
	}

	if card.CardStatus != domain.CardStatusActive {
		return AuthorizationResult{Decision: AuthorizationDecisionDecline, Reason: "CARD_NOT_ACTIVE"}, nil
	}

	mKey := normalizeMerchantKey(event.MerchantName)
	if mKey != "" {
		sub, err := uc.subRepo.GetByCardAndMerchantKey(ctx, card.ID, mKey)
		if err == nil && sub != nil && sub.IsBlocked {
			return AuthorizationResult{Decision: AuthorizationDecisionDecline, Reason: "SUBSCRIPTION_BLOCKED"}, nil
		}
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return AuthorizationResult{}, wrapper.Wrap(err)
		}
	}

	amt, err := decimal.NewFromString(strings.TrimSpace(event.Amount))
	if err != nil {
		return AuthorizationResult{}, domain.NewInvalidInput("amount invalid")
	}

	details := strings.TrimSpace(event.MerchantName)
	err = uc.cardUC.SpendFromCardWithDetails(ctx, card.UserID, card.ID, amt, details)
	if err != nil {
		if decline, ok := authorizationDeclineFromSpendErr(err); ok {
			return decline, nil
		}

		return AuthorizationResult{}, wrapper.Wrap(err)
	}

	executedAt := time.Now().UTC()
	if strings.TrimSpace(event.ExecutedAt) != "" {
		t, tErr := time.Parse(time.RFC3339, event.ExecutedAt)
		if tErr == nil {
			executedAt = t.UTC()
		}
	}

	_, err = uc.subRepo.UpsertOnCharge(ctx, card.UserID, card.ID, event.MerchantName, amt, event.Currency, executedAt)
	if err != nil {
		return AuthorizationResult{}, wrapper.Wrap(err)
	}

	return AuthorizationResult{Decision: AuthorizationDecisionApprove}, nil
}

func authorizationDeclineFromSpendErr(err error) (AuthorizationResult, bool) {
	if err == nil {
		return AuthorizationResult{}, false
	}

	if errors.Is(err, domain.ErrInsufficientFunds) {
		return AuthorizationResult{Decision: AuthorizationDecisionDecline, Reason: "INSUFFICIENT_FUNDS"}, true
	}

	if !errors.Is(err, domain.ErrInvalidInput) {
		return AuthorizationResult{}, false
	}

	return AuthorizationResult{
		Decision: AuthorizationDecisionDecline,
		Reason:   spendInvalidInputDeclineReason(err),
	}, true
}

func spendInvalidInputDeclineReason(err error) string {
	switch {
	case domain.IsInvalidInputCode(err, "daily spend limit exceeded"):
		return "DAILY_LIMIT_EXCEEDED"
	case domain.IsInvalidInputCode(err, "monthly spend limit exceeded"):
		return "MONTHLY_LIMIT_EXCEEDED"
	case domain.IsInvalidInputCode(err, "card is not active"):
		return "CARD_NOT_ACTIVE"
	case domain.IsInvalidInputCode(err, "card not found"):
		return "CARD_NOT_FOUND"
	case domain.IsInvalidInputCode(err, "amount must be positive"):
		return "INVALID_AMOUNT"
	default:
		return "DECLINED"
	}
}
