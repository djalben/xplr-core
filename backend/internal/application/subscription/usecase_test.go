package subscription_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/application/card/mocks"
	"github.com/djalben/xplr-core/backend/internal/application/subscription"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

type fakeSubRepo struct {
	getFn    func(ctx context.Context, cardID domain.UUID, merchantKey string) (*domain.CardSubscription, error)
	upsertFn func(ctx context.Context, userID, cardID domain.UUID, merchantName string, amount domain.Numeric, currency string) error
}

func (f *fakeSubRepo) UpsertOnCharge(ctx context.Context, userID domain.UUID, cardID domain.UUID, merchantName string, amount domain.Numeric, currency string, _ time.Time) (*domain.CardSubscription, error) {
	if f.upsertFn != nil {
		err := f.upsertFn(ctx, userID, cardID, merchantName, amount, currency)
		if err != nil {
			return nil, err
		}
	}
	return &domain.CardSubscription{ID: uuid.New(), UserID: userID, CardID: cardID, MerchantName: merchantName, MerchantKey: merchantName, ChargeCount: 1}, nil
}
func (f *fakeSubRepo) ListByUserID(context.Context, domain.UUID) ([]*domain.CardSubscription, error) { return nil, nil }
func (f *fakeSubRepo) SetBlocked(context.Context, domain.UUID, domain.UUID, bool) error                 { return nil }
func (f *fakeSubRepo) SetBlockedByCardID(context.Context, domain.UUID, domain.UUID, bool) error        { return nil }
func (f *fakeSubRepo) GetByCardAndMerchantKey(ctx context.Context, cardID domain.UUID, merchantKey string) (*domain.CardSubscription, error) {
	if f.getFn == nil {
		return nil, domain.NewNotFound("subscription not found")
	}
	return f.getFn(ctx, cardID, merchantKey)
}

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

	subRepo := &fakeSubRepo{
		getFn: func(ctx context.Context, cardID domain.UUID, merchantKey string) (*domain.CardSubscription, error) {
			return &domain.CardSubscription{ID: uuid.New(), CardID: cardID, MerchantKey: merchantKey, IsBlocked: true}, nil
		},
	}

	cardUC := card.NewUseCase(cardRepo, walletRepo, txRepo, gradeRepo, nil)
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

	upsertCalled := false
	subRepo := &fakeSubRepo{
		getFn: func(ctx context.Context, cardID domain.UUID, merchantKey string) (*domain.CardSubscription, error) {
			return nil, domain.NewNotFound("subscription not found")
		},
		upsertFn: func(ctx context.Context, userID, cardID domain.UUID, merchantName string, amount domain.Numeric, currency string) error {
			upsertCalled = true
			if merchantName != "Netflix" || currency != "USD" {
				return errors.New("bad args")
			}
			return nil
		},
	}

	cardUC := card.NewUseCase(cardRepo, walletRepo, txRepo, gradeRepo, nil)
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
	if !upsertCalled {
		t.Fatalf("expected upsert")
	}
}

