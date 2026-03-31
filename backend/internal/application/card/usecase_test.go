package card_test

//go:generate GOFLAGS=-mod=mod go run github.com/golang/mock/mockgen@v1.6.0 -destination=./mocks/ports_mock.go -package=mocks github.com/djalben/xplr-core/backend/internal/ports CardRepository,WalletRepository,TransactionRepository,GradeRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/application/card/mocks"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func testUID() domain.UUID {
	return uuid.MustParse("11111111-1111-1111-1111-111111111111")
}

func testCardID() domain.UUID {
	return uuid.MustParse("22222222-2222-2222-2222-222222222222")
}

func activeCard(uid domain.UUID, ct domain.CardType) *domain.Card {
	c, err := domain.NewCard(uid, ct, "p", "n")
	if err != nil {
		panic(err)
	}

	return c
}

// cardUCTest — моки и use case без лишних возвращаемых значений (dogsled/unparam).
type cardUCTest struct {
	UC     *card.UseCase
	Cards  *mocks.MockCardRepository
	Tx     *mocks.MockTransactionRepository
	Grades *mocks.MockGradeRepository
}

func newCardUCTest(ctrl *gomock.Controller) cardUCTest {
	cr := mocks.NewMockCardRepository(ctrl)
	wr := mocks.NewMockWalletRepository(ctrl)
	tr := mocks.NewMockTransactionRepository(ctrl)
	gr := mocks.NewMockGradeRepository(ctrl)
	uc := card.NewUseCase(cr, wr, tr, gr, nil)

	return cardUCTest{UC: uc, Cards: cr, Tx: tr, Grades: gr}
}

func TestUseCase_BuyCard_StandardLimitReached(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	ct := domain.CardTypeSubscriptions

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	m.Grades.EXPECT().EnsureGrade(gomock.Any(), uid).Return(nil)
	m.Grades.EXPECT().GetByUserID(gomock.Any(), uid).Return(&domain.UserGrade{Grade: domain.UserGradeStandard}, nil)

	list := []*domain.Card{activeCard(uid, ct), activeCard(uid, ct), activeCard(uid, ct)}
	m.Cards.EXPECT().ListByUserID(gomock.Any(), uid).Return(list, nil)

	_, err := m.UC.BuyCard(ctx, uid, ct, "x")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestUseCase_BuyCard_OK_EmptyList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	ct := domain.CardTypeTravel

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	m.Grades.EXPECT().EnsureGrade(gomock.Any(), uid).Return(nil)
	m.Grades.EXPECT().GetByUserID(gomock.Any(), uid).Return(&domain.UserGrade{Grade: domain.UserGradeGold}, nil)
	m.Cards.EXPECT().ListByUserID(gomock.Any(), uid).Return([]*domain.Card{}, nil)
	m.Cards.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
	m.Tx.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	got, err := m.UC.BuyCard(ctx, uid, ct, "nick")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.UserID != uid || got.CardType != ct {
		t.Fatalf("card: %+v", got)
	}
}

func TestUseCase_SpendFromCard_InvalidAmount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	err := m.UC.SpendFromCard(context.Background(), testUID(), testCardID(), domain.NewNumeric(0))
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

func TestUseCase_SpendFromCard_NotActive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, "p", "n")
	c.ID = cid
	c.CardStatus = domain.CardStatusBlocked

	m.Cards.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)

	err := m.UC.SpendFromCard(ctx, uid, cid, domain.NewNumeric(1))
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

func TestUseCase_SpendFromCard_WrongUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	other := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	c, _ := domain.NewCard(other, domain.CardTypeSubscriptions, "p", "n")
	c.ID = cid

	m.Cards.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)

	err := m.UC.SpendFromCard(ctx, uid, cid, domain.NewNumeric(1))
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

func TestUseCase_SpendFromCard_DailyLimitExceeded(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, "p", "n")
	c.ID = cid
	c.DailySpendLimit = domain.NewNumeric(100)

	m.Cards.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)
	m.Tx.EXPECT().SumCardSpendByCardID(gomock.Any(), cid, gomock.Any(), gomock.Any()).Return(domain.NewNumeric(95), nil)

	err := m.UC.SpendFromCard(ctx, uid, cid, domain.NewNumeric(10))
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

func TestUseCase_SpendFromCard_MonthlyLimitExceeded(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, "p", "n")
	c.ID = cid
	c.Balance = domain.NewNumeric(10000)
	// subscriptions monthly cap 5000
	m.Cards.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)
	m.Tx.EXPECT().SumCardSpendByCardID(gomock.Any(), cid, gomock.Any(), gomock.Any()).Return(domain.NewNumeric(0), nil)
	m.Tx.EXPECT().SumCardSpendByUserAndCardType(gomock.Any(), uid, domain.CardTypeSubscriptions, gomock.Any(), gomock.Any()).
		Return(domain.NewNumeric(4990), nil)

	err := m.UC.SpendFromCard(ctx, uid, cid, domain.NewNumeric(20))
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

func TestUseCase_SpendFromCard_InsufficientBalance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, "p", "n")
	c.ID = cid
	c.Balance = domain.NewNumeric(1)

	m.Cards.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)
	m.Tx.EXPECT().SumCardSpendByCardID(gomock.Any(), cid, gomock.Any(), gomock.Any()).Return(domain.NewNumeric(0), nil)
	m.Tx.EXPECT().SumCardSpendByUserAndCardType(gomock.Any(), uid, domain.CardTypeSubscriptions, gomock.Any(), gomock.Any()).
		Return(domain.NewNumeric(0), nil)

	err := m.UC.SpendFromCard(ctx, uid, cid, domain.NewNumeric(5))
	if err == nil || !errors.Is(err, domain.ErrInsufficientFunds) {
		t.Fatalf("got %v", err)
	}
}

func TestUseCase_SpendFromCard_OK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, "p", "n")
	c.ID = cid
	c.Balance = domain.NewNumeric(50)
	c.DailySpendLimit = domain.NewNumeric(1000)

	m.Cards.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)
	m.Tx.EXPECT().SumCardSpendByCardID(gomock.Any(), cid, gomock.Any(), gomock.Any()).Return(domain.NewNumeric(10), nil)
	m.Tx.EXPECT().SumCardSpendByUserAndCardType(gomock.Any(), uid, domain.CardTypeSubscriptions, gomock.Any(), gomock.Any()).
		Return(domain.NewNumeric(100), nil)
	m.Cards.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, upd *domain.Card) error {
		if !upd.Balance.Equal(domain.NewNumeric(40)) {
			t.Fatalf("balance after spend: %s", upd.Balance.String())
		}

		return nil
	})
	m.Tx.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	err := m.UC.SpendFromCard(ctx, uid, cid, domain.NewNumeric(10))
	if err != nil {
		t.Fatal(err)
	}
}
