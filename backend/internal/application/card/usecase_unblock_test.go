package card_test

import (
	"context"
	"errors"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang/mock/gomock"
)

func TestUseCase_UnblockCard_ClosedRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, "p", "n")
	c.ID = cid
	c.CardStatus = domain.CardStatusClosed

	m.Cards.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)

	err := m.UC.UnblockCard(ctx, uid, cid)
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

func TestUseCase_UpdateStatus_ActiveOnClosedRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := testUID()
	cid := testCardID()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	m := newCardUCTest(ctrl)

	c, _ := domain.NewCard(uid, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, "p", "n")
	c.ID = cid
	c.CardStatus = domain.CardStatusClosed

	m.Cards.EXPECT().GetByID(gomock.Any(), cid).Return(c, nil)

	err := m.UC.UpdateStatus(ctx, uid, cid, domain.CardStatusActive)
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}
