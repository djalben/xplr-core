package card_test

import (
	"context"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/application/card/mocks"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang/mock/gomock"
)

func TestUseCase_CloseCard_RefundsAndZerosBalance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cr := mocks.NewMockCardRepository(ctrl)
	wr := mocks.NewMockWalletRepository(ctrl)
	tr := mocks.NewMockTransactionRepository(ctrl)
	gr := mocks.NewMockGradeRepository(ctrl)
	uc := card.NewUseCase(cr, wr, tr, gr, nil)

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, "p", "n")
	c.ID = cid
	c.Balance = domain.NewNumeric(25)

	w := domain.NewWallet(uid)
	_ = w.TopUp(domain.NewNumeric(10))

	cr.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)
	cr.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, upd *domain.Card) error {
		if upd.CardStatus != domain.CardStatusClosed {
			t.Fatalf("status: %s", upd.CardStatus)
		}
		if !upd.Balance.Equal(domain.NewNumeric(0)) {
			t.Fatalf("card balance after close: %s", upd.Balance.String())
		}

		return nil
	})
	wr.EXPECT().GetByUserID(gomock.Any(), uid).Return(w, nil)
	wr.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, upd *domain.Wallet) error {
		if !upd.Balance.Equal(domain.NewNumeric(35)) {
			t.Fatalf("wallet balance after refund: %s", upd.Balance.String())
		}

		return nil
	})
	tr.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, tx *domain.Transaction) error {
		if tx.TransactionType != domain.TransactionTypeCardRefund {
			t.Fatalf("tx type: %s", tx.TransactionType)
		}
		if tx.CardID == nil || *tx.CardID != cid {
			t.Fatal("tx card_id mismatch")
		}
		if !tx.Amount.Equal(domain.NewNumeric(25)) {
			t.Fatalf("tx amount: %s", tx.Amount.String())
		}

		return nil
	})

	err := uc.CloseCard(ctx, uid, cid)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUseCase_CloseCard_IdempotentWhenAlreadyClosed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cr := mocks.NewMockCardRepository(ctrl)
	wr := mocks.NewMockWalletRepository(ctrl)
	tr := mocks.NewMockTransactionRepository(ctrl)
	gr := mocks.NewMockGradeRepository(ctrl)
	uc := card.NewUseCase(cr, wr, tr, gr, nil)

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, "p", "n")
	c.ID = cid
	c.CardStatus = domain.CardStatusClosed
	c.Balance = domain.NewNumeric(50)

	cr.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)

	err := uc.CloseCard(ctx, uid, cid)
	if err != nil {
		t.Fatal(err)
	}
}
