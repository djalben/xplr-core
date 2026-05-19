package card_test

import (
	"context"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/application/card/mocks"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang/mock/gomock"
)

func TestUseCase_AutoTopUpCard_OK(t *testing.T) {
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

	w := domain.NewWallet(uid)
	_ = w.TopUp(domain.NewNumeric(100))
	w.AutoTopUpEnabled = true

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, "p", "n")
	c.ID = cid
	c.Balance = domain.NewNumeric(5)

	wr.EXPECT().GetByUserID(gomock.Any(), uid).Return(w, nil)
	cr.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)
	wr.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, upd *domain.Wallet) error {
		if !upd.Balance.Equal(domain.NewNumeric(90)) {
			t.Fatalf("wallet balance after withdraw: %s", upd.Balance.String())
		}

		return nil
	})
	cr.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, upd *domain.Card) error {
		if !upd.Balance.Equal(domain.NewNumeric(15)) {
			t.Fatalf("card balance after top-up: %s", upd.Balance.String())
		}

		return nil
	})
	tr.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	err := uc.AutoTopUpCard(ctx, uid, cid, domain.NewNumeric(10))
	if err != nil {
		t.Fatal(err)
	}
}
