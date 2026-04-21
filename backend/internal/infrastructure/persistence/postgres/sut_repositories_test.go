package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/infrastructure/persistence/postgres"
)

func TestUserRepository_SaveGetByEmailUpdate(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	repo := postgres.NewUserRepository(db)
	email := uniqueEmail(t)
	u := newTestUser(t, db, email)

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Email != email {
		t.Fatalf("email: got %q want %q", got.Email, email)
	}

	byMail, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if byMail.ID != u.ID {
		t.Fatalf("GetByEmail id mismatch")
	}

	u.KYCStatus = domain.KYCApproved
	err = repo.Update(ctx, u)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	again, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if again.KYCStatus != domain.KYCApproved {
		t.Fatalf("kyc: got %s want %s", again.KYCStatus, domain.KYCApproved)
	}

	_, err = repo.GetByID(ctx, domain.NewUUID())
	if err == nil {
		t.Fatal("GetByID missing row: want error")
	}
}

func TestWalletRepository_EnsureGetUpdate(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	u := newTestUser(t, db, uniqueEmail(t))
	wrepo := postgres.NewWalletRepository(db)

	err := wrepo.EnsureWallet(ctx, u.ID)
	if err != nil {
		t.Fatalf("EnsureWallet: %v", err)
	}
	err = wrepo.EnsureWallet(ctx, u.ID)
	if err != nil {
		t.Fatalf("EnsureWallet idempotent: %v", err)
	}

	w, err := wrepo.GetByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if w.UserID != u.ID {
		t.Fatalf("wallet user_id")
	}
	if !w.Balance.Equal(domain.NewNumeric(0)) {
		t.Fatalf("balance: got %s", w.Balance.String())
	}

	w.Balance = domain.NewNumeric(100.5)
	w.AutoTopUpEnabled = true
	err = wrepo.Update(ctx, w)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	w2, err := wrepo.GetByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByUserID 2: %v", err)
	}
	if !w2.Balance.Equal(domain.NewNumeric(100.5)) {
		t.Fatalf("balance after update: %s", w2.Balance.String())
	}
	if !w2.AutoTopUpEnabled {
		t.Fatal("auto_topup")
	}
}

func TestGradeRepository_EnsureGetUpdate(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	u := newTestUser(t, db, uniqueEmail(t))
	grepo := postgres.NewGradeRepository(db)

	err := grepo.EnsureGrade(ctx, u.ID)
	if err != nil {
		t.Fatalf("EnsureGrade: %v", err)
	}

	g, err := grepo.GetByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if g.Grade != "STANDARD" {
		t.Fatalf("grade: %s", g.Grade)
	}

	g.Grade = "GOLD"
	g.TotalSpent = domain.NewNumeric(50)
	g.FeePercent = domain.NewNumeric(5.5)
	err = grepo.Update(ctx, g)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	g2, err := grepo.GetByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByUserID 2: %v", err)
	}
	if g2.Grade != "GOLD" {
		t.Fatalf("grade after update: %s", g2.Grade)
	}
}

func TestCardRepository_SaveGetListUpdate(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	u := newTestUser(t, db, uniqueEmail(t))
	card, err := domain.NewCard(u.ID, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, "sut-prov", "c1")
	if err != nil {
		t.Fatalf("NewCard: %v", err)
	}

	repo := postgres.NewCardRepository(db)
	err = repo.Save(ctx, card)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.GetByID(ctx, card.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.UserID != u.ID || got.Last4Digits != card.Last4Digits {
		t.Fatalf("GetByID mismatch")
	}

	list, err := repo.ListByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list len: %d", len(list))
	}

	got.Balance = domain.NewNumeric(25)
	got.CardStatus = "ACTIVE"
	got.Nickname = "n2"
	err = repo.Update(ctx, got)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	again, err := repo.GetByID(ctx, card.ID)
	if err != nil {
		t.Fatalf("GetByID 2: %v", err)
	}
	if !again.Balance.Equal(domain.NewNumeric(25)) {
		t.Fatalf("balance: %s", again.Balance.String())
	}
	if again.Nickname != "n2" {
		t.Fatalf("nickname: %s", again.Nickname)
	}
}

func TestTransactionRepository_SaveAndQueries(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	u := newTestUser(t, db, uniqueEmail(t))
	card, err := domain.NewCard(u.ID, domain.CardTypeTravel, domain.CardCurrencyUSD, "p2", "c2")
	if err != nil {
		t.Fatalf("NewCard: %v", err)
	}
	cardRepo := postgres.NewCardRepository(db)
	err = cardRepo.Save(ctx, card)
	if err != nil {
		t.Fatalf("card save: %v", err)
	}

	t0 := time.Date(2025, 3, 10, 12, 0, 0, 0, time.UTC)
	from := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC)

	txWallet := domain.NewTransaction(u.ID, nil, domain.NewNumeric(10), domain.NewNumeric(0),
		"TOPUP_WALLET", "COMPLETED", "w")
	txWallet.ExecutedAt = t0

	txCard := domain.NewTransaction(u.ID, &card.ID, domain.NewNumeric(5), domain.NewNumeric(0),
		"TOPUP_CARD", "COMPLETED", "c")
	txCard.ExecutedAt = t0

	repo := postgres.NewTransactionRepository(db)
	err = repo.Save(ctx, txWallet)
	if err != nil {
		t.Fatalf("Save wallet tx: %v", err)
	}
	err = repo.Save(ctx, txCard)
	if err != nil {
		t.Fatalf("Save card tx: %v", err)
	}

	wList, err := repo.GetWalletTransactions(ctx, u.ID, from, to)
	if err != nil {
		t.Fatalf("GetWalletTransactions: %v", err)
	}
	if len(wList) != 1 {
		t.Fatalf("wallet txs: %d", len(wList))
	}
	if wList[0].CardID != nil {
		t.Fatal("wallet tx must have card_id NULL")
	}

	cList, err := repo.GetCardTransactions(ctx, card.ID, from, to)
	if err != nil {
		t.Fatalf("GetCardTransactions: %v", err)
	}
	if len(cList) != 1 {
		t.Fatalf("card txs: %d", len(cList))
	}

	all, err := repo.GetByUserID(ctx, u.ID, from, to, 10)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("unified: %d", len(all))
	}
}

func TestTransactionRepository_SumCardSpend(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	u := newTestUser(t, db, uniqueEmail(t))

	cardSub, err := domain.NewCard(u.ID, domain.CardTypeSubscriptions, domain.CardCurrencyUSD, "p1", "n1")
	if err != nil {
		t.Fatalf("NewCard: %v", err)
	}

	cardTr, err := domain.NewCard(u.ID, domain.CardTypeTravel, domain.CardCurrencyUSD, "p2", "n2")
	if err != nil {
		t.Fatalf("NewCard: %v", err)
	}

	cardRepo := postgres.NewCardRepository(db)
	err = cardRepo.Save(ctx, cardSub)
	if err != nil {
		t.Fatalf("save sub: %v", err)
	}
	err = cardRepo.Save(ctx, cardTr)
	if err != nil {
		t.Fatalf("save travel: %v", err)
	}

	tMarch := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)
	mStart := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	mEnd := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)

	repo := postgres.NewTransactionRepository(db)

	sp1 := domain.NewTransaction(u.ID, &cardSub.ID, domain.NewNumeric(100), domain.NewNumeric(0),
		domain.TransactionTypeCardSpend, "COMPLETED", "a")
	sp1.ExecutedAt = tMarch

	sp2 := domain.NewTransaction(u.ID, &cardSub.ID, domain.NewNumeric(50), domain.NewNumeric(0),
		domain.TransactionTypeCardSpend, "COMPLETED", "b")
	sp2.ExecutedAt = tMarch

	spTr := domain.NewTransaction(u.ID, &cardTr.ID, domain.NewNumeric(10), domain.NewNumeric(0),
		domain.TransactionTypeCardSpend, "COMPLETED", "c")
	spTr.ExecutedAt = tMarch

	noise := domain.NewTransaction(u.ID, &cardSub.ID, domain.NewNumeric(999), domain.NewNumeric(0),
		"TOPUP_CARD", "COMPLETED", "noise")
	noise.ExecutedAt = tMarch

	for _, tx := range []*domain.Transaction{sp1, sp2, spTr, noise} {
		err = repo.Save(ctx, tx)
		if err != nil {
			t.Fatalf("Save tx: %v", err)
		}
	}

	sumCard, err := repo.SumCardSpendByCardID(ctx, cardSub.ID, mStart, mEnd)
	if err != nil {
		t.Fatalf("SumCardSpendByCardID: %v", err)
	}
	if !sumCard.Equal(domain.NewNumeric(150)) {
		t.Fatalf("sum by card: got %s want 150", sumCard.String())
	}

	sumSubType, err := repo.SumCardSpendByUserAndCardType(ctx, u.ID, domain.CardTypeSubscriptions, mStart, mEnd)
	if err != nil {
		t.Fatalf("SumCardSpendByUserAndCardType subscriptions: %v", err)
	}
	if !sumSubType.Equal(domain.NewNumeric(150)) {
		t.Fatalf("sum subscriptions type: got %s want 150", sumSubType.String())
	}

	sumTravelType, err := repo.SumCardSpendByUserAndCardType(ctx, u.ID, domain.CardTypeTravel, mStart, mEnd)
	if err != nil {
		t.Fatalf("SumCardSpendByUserAndCardType travel: %v", err)
	}
	if !sumTravelType.Equal(domain.NewNumeric(10)) {
		t.Fatalf("sum travel type: got %s want 10", sumTravelType.String())
	}

	dayStart, dayEnd := domain.DayBoundsUTC(tMarch)
	sumDay, err := repo.SumCardSpendByCardID(ctx, cardSub.ID, dayStart, dayEnd)
	if err != nil {
		t.Fatalf("SumCardSpendByCardID day: %v", err)
	}
	if !sumDay.Equal(domain.NewNumeric(150)) {
		t.Fatalf("sum same day: got %s want 150", sumDay.String())
	}
}

func TestTicketRepository_SaveGetUpdate(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	u := newTestUser(t, db, uniqueEmail(t))
	tk := domain.NewTicket(u.ID, "subj", "msg", nil)

	repo := postgres.NewTicketRepository(db)
	err := repo.Save(ctx, tk)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.GetByID(ctx, tk.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.UserMessage != "msg" {
		t.Fatalf("message: %s", got.UserMessage)
	}

	admin := domain.NewUUID()
	got.Take(admin)
	err = repo.Update(ctx, got)
	if err != nil {
		t.Fatalf("Update take: %v", err)
	}

	got2, err := repo.GetByID(ctx, tk.ID)
	if err != nil {
		t.Fatalf("GetByID 2: %v", err)
	}
	if got2.Status != domain.TicketInProgress {
		t.Fatalf("status: %s", got2.Status)
	}
}

func TestReferralRepository_CountAndSum(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	a := newTestUser(t, db, uniqueEmail(t))
	b := newTestUser(t, db, uniqueEmail(t))
	c := newTestUser(t, db, uniqueEmail(t))

	_, err := db.ExecContext(ctx,
		`INSERT INTO referrals (referrer_id, referred_id, status, commission_earned) VALUES ($1, $2, 'PENDING', $3)`,
		a.ID, b.ID, 10.0)
	if err != nil {
		t.Fatalf("insert referral: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO referrals (referrer_id, referred_id, status, commission_earned) VALUES ($1, $2, 'PENDING', $3)`,
		a.ID, c.ID, 5.5)
	if err != nil {
		t.Fatalf("insert referral 2: %v", err)
	}

	repo := postgres.NewReferralRepository(db)
	n, err := repo.CountByReferrer(ctx, a.ID)
	if err != nil {
		t.Fatalf("CountByReferrer: %v", err)
	}
	if n != 2 {
		t.Fatalf("count: %d want 2", n)
	}

	sum, err := repo.TotalEarningsByReferrer(ctx, a.ID)
	if err != nil {
		t.Fatalf("TotalEarningsByReferrer: %v", err)
	}
	if !sum.Equal(domain.NewNumeric(15.5)) {
		t.Fatalf("sum: %s", sum.String())
	}
}

func TestCommissionRepository_GetListUpdate(t *testing.T) {
	db := openTestDB(t)
	// commission_config не трогаем truncate — там сиды; откатываем значение в defer

	ctx := context.Background()
	repo := postgres.NewCommissionConfigRepository(db)

	key := domain.FeeStandard
	cfg, err := repo.GetByKey(ctx, key)
	if err != nil {
		t.Fatalf("GetByKey: %v", err)
	}
	origVal := cfg.Value
	origDesc := cfg.Description

	t.Cleanup(func() {
		cfg2, err := repo.GetByKey(context.Background(), key)
		if err != nil {
			t.Fatalf("cleanup GetByKey: %v", err)
		}
		cfg2.Value = origVal
		cfg2.Description = origDesc
		err = repo.Update(context.Background(), cfg2)
		if err != nil {
			t.Fatalf("cleanup Update: %v", err)
		}
	})

	cfg.Value = domain.NewNumeric(6.71)
	cfg.Description = "SUT"
	err = repo.Update(ctx, cfg)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByKey(ctx, key)
	if err != nil {
		t.Fatalf("GetByKey after update: %v", err)
	}
	if !got.Value.Equal(domain.NewNumeric(6.71)) {
		t.Fatalf("value: %s", got.Value.String())
	}

	all, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) < 1 {
		t.Fatalf("ListAll empty")
	}

	_, err = repo.GetByKey(ctx, "no_such_key_ever")
	if err == nil {
		t.Fatal("GetByKey missing: want error")
	}
}
