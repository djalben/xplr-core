package card_test

import (
	"context"
	"errors"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/application/card/mocks"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang/mock/gomock"
)

func TestUseCase_BuyCard_InsufficientFunds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	ct := domain.CardTypeSubscriptions

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cr := mocks.NewMockCardRepository(ctrl)
	wr := mocks.NewMockWalletRepository(ctrl)
	tr := mocks.NewMockTransactionRepository(ctrl)
	gr := mocks.NewMockGradeRepository(ctrl)
	uc := card.NewUseCase(cr, wr, tr, gr, nil)

	gr.EXPECT().EnsureGrade(gomock.Any(), uid).Return(nil)
	gr.EXPECT().GetByUserID(gomock.Any(), uid).Return(&domain.UserGrade{Grade: domain.UserGradeGold}, nil)
	cr.EXPECT().ListByUserID(gomock.Any(), uid).Return([]*domain.Card{}, nil)

	w := domain.NewWallet(uid)
	_ = w.TopUp(domain.NewNumeric(1))
	wr.EXPECT().GetByUserID(gomock.Any(), uid).Return(w, nil)

	_, err := uc.BuyCard(ctx, uid, ct, "nick", domain.CardCurrencyUSD)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domain.ErrInsufficientFunds) {
		t.Fatalf("want ErrInsufficientFunds, got %v", err)
	}
}
