package subscription_test

import (
	"context"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/application/card/mocks"
	"github.com/djalben/xplr-core/backend/internal/application/subscription"
	submocks "github.com/djalben/xplr-core/backend/internal/application/subscription/mocks"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func TestUseCase_HandleAuthorization_BlockedDecline(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cardRepo := mocks.NewMockCardRepository(ctrl)
	walletRepo := mocks.NewMockWalletRepository(ctrl)
	txRepo := mocks.NewMockTransactionRepository(ctrl)
	gradeRepo := mocks.NewMockGradeRepository(ctrl)

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	providerCardID := "prov-123"

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, providerCardID, "n")
	c.ID = cid
	c.CardStatus = domain.CardStatusActive
	c.Balance = domain.NewNumeric(100)

	cardRepo.EXPECT().GetByProviderCardID(gomock.Any(), providerCardID).Return(c, nil)

	subRepo := submocks.NewMockCardSubscriptionRepository(ctrl)
	subRepo.EXPECT().
		GetByCardAndMerchantKey(gomock.Any(), cid, "netflix").
		Return(&domain.CardSubscription{ID: uuid.New(), CardID: cid, MerchantKey: "netflix", IsBlocked: true}, nil)

	cardUC := card.NewUseCase(cardRepo, walletRepo, txRepo, gradeRepo)
	uc := subscription.NewUseCase(cardRepo, subRepo, cardUC)

	res, err := uc.HandleAuthorization(ctx, subscription.AuthorizationEvent{
		ProviderCardID: providerCardID,
		Amount:         "10.00",
		Currency:       "USD",
		MerchantName:   "Netflix",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Decision != subscription.AuthorizationDecisionDecline || res.Reason != "SUBSCRIPTION_BLOCKED" {
		t.Fatalf("res: %+v", res)
	}
}

func TestUseCase_HandleAuthorization_ApproveAndUpsert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cardRepo := mocks.NewMockCardRepository(ctrl)
	walletRepo := mocks.NewMockWalletRepository(ctrl)
	txRepo := mocks.NewMockTransactionRepository(ctrl)
	gradeRepo := mocks.NewMockGradeRepository(ctrl)

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	providerCardID := "prov-123"

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, providerCardID, "n")
	c.ID = cid
	c.CardStatus = domain.CardStatusActive
	c.Balance = domain.NewNumeric(100)

	cardRepo.EXPECT().GetByProviderCardID(gomock.Any(), providerCardID).Return(c, nil)
	cardRepo.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)
	txRepo.EXPECT().SumCardSpendByCardID(gomock.Any(), cid, gomock.Any(), gomock.Any()).Return(domain.NewNumeric(0), nil)
	txRepo.EXPECT().SumCardSpendByUserAndCardType(gomock.Any(), uid, domain.CardTypeSubscriptions, gomock.Any(), gomock.Any()).
		Return(domain.NewNumeric(0), nil)
	cardRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	txRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	subRepo := submocks.NewMockCardSubscriptionRepository(ctrl)
	subRepo.EXPECT().
		GetByCardAndMerchantKey(gomock.Any(), cid, "netflix").
		Return(nil, domain.NewNotFound("subscription not found"))
	subRepo.EXPECT().
		UpsertOnCharge(gomock.Any(), uid, cid, "Netflix", gomock.Any(), "USD", gomock.Any()).
		Return(&domain.CardSubscription{ID: uuid.New(), UserID: uid, CardID: cid, MerchantName: "Netflix", MerchantKey: "netflix", ChargeCount: 1}, nil)

	cardUC := card.NewUseCase(cardRepo, walletRepo, txRepo, gradeRepo)
	uc := subscription.NewUseCase(cardRepo, subRepo, cardUC)

	res, err := uc.HandleAuthorization(ctx, subscription.AuthorizationEvent{
		ProviderCardID: providerCardID,
		Amount:         "10.00",
		Currency:       "USD",
		MerchantName:   "Netflix",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Decision != subscription.AuthorizationDecisionApprove {
		t.Fatalf("res: %+v", res)
	}
}
