package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/infrastructure/persistence/postgres"
)

func TestSubscriptionRepository_UpsertListBlock(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	u := newTestUser(t, db, uniqueEmail(t))

	card, err := domain.NewCard(u.ID, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, "provider-1", "n1")
	if err != nil {
		t.Fatalf("NewCard: %v", err)
	}

	cardRepo := postgres.NewCardRepository(db)
	err = cardRepo.Save(ctx, card)
	if err != nil {
		t.Fatalf("card save: %v", err)
	}

	repo := postgres.NewCardSubscriptionRepository(db)

	now := time.Now().UTC()
	one, err := repo.UpsertOnCharge(ctx, u.ID, card.ID, "Netflix", domain.NewNumeric(9.99), "USD", now)
	if err != nil {
		t.Fatalf("UpsertOnCharge: %v", err)
	}
	if one == nil || one.ChargeCount != 1 || one.MerchantKey == "" {
		t.Fatalf("bad sub: %+v", one)
	}

	two, err := repo.UpsertOnCharge(ctx, u.ID, card.ID, "Netflix", domain.NewNumeric(10.99), "USD", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("UpsertOnCharge #2: %v", err)
	}
	if two == nil || two.ChargeCount != 2 {
		t.Fatalf("charge_count: %+v", two)
	}

	list, err := repo.ListByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len: %d", len(list))
	}

	err = repo.SetBlocked(ctx, u.ID, list[0].ID, true)
	if err != nil {
		t.Fatalf("SetBlocked: %v", err)
	}

	got, err := repo.GetByCardAndMerchantKey(ctx, card.ID, list[0].MerchantKey)
	if err != nil {
		t.Fatalf("GetByCardAndMerchantKey: %v", err)
	}
	if got == nil || got.IsBlocked != true || got.BlockedAt == nil {
		t.Fatalf("blocked: %+v", got)
	}

	err = repo.SetBlockedByCardID(ctx, u.ID, card.ID, false)
	if err != nil {
		t.Fatalf("SetBlockedByCardID: %v", err)
	}

	got2, err := repo.GetByCardAndMerchantKey(ctx, card.ID, list[0].MerchantKey)
	if err != nil {
		t.Fatalf("GetByCardAndMerchantKey #2: %v", err)
	}
	if got2 == nil || got2.IsBlocked != false || got2.BlockedAt != nil {
		t.Fatalf("unblocked: %+v", got2)
	}
}

