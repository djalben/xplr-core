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
	if err := repo.Update(ctx, u); err != nil {
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

	if err := wrepo.EnsureWallet(ctx, u.ID); err != nil {
		t.Fatalf("EnsureWallet: %v", err)
	}
	if err := wrepo.EnsureWallet(ctx, u.ID); err != nil {
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
	if err := wrepo.Update(ctx, w); err != nil {
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

	if err := grepo.EnsureGrade(ctx, u.ID); err != nil {
		t.Fatalf("EnsureGrade: %v", err)
	}

	g, err := grepo.GetByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if g.Grade != "STANDARD" {
		t.Fatalf("grade: %s", g.Grade)
	}

	g.Grade = "SILVER"
	g.TotalSpent = domain.NewNumeric(50)
	g.FeePercent = domain.NewNumeric(5.5)
	if err := grepo.Update(ctx, g); err != nil {
		t.Fatalf("Update: %v", err)
	}

	g2, err := grepo.GetByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByUserID 2: %v", err)
	}
	if g2.Grade != "SILVER" {
		t.Fatalf("grade after update: %s", g2.Grade)
	}
}

func TestCardRepository_SaveGetListUpdate(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	u := newTestUser(t, db, uniqueEmail(t))
	card, err := domain.NewCard(u.ID, domain.CardTypeSubscriptions, "sut-prov", "c1")
	if err != nil {
		t.Fatalf("NewCard: %v", err)
	}

	repo := postgres.NewCardRepository(db)
	if err := repo.Save(ctx, card); err != nil {
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
	if err := repo.Update(ctx, got); err != nil {
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
	card, err := domain.NewCard(u.ID, domain.CardTypeTravel, "p2", "c2")
	if err != nil {
		t.Fatalf("NewCard: %v", err)
	}
	cardRepo := postgres.NewCardRepository(db)
	if err := cardRepo.Save(ctx, card); err != nil {
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
	if err := repo.Save(ctx, txWallet); err != nil {
		t.Fatalf("Save wallet tx: %v", err)
	}
	if err := repo.Save(ctx, txCard); err != nil {
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

func TestTicketRepository_SaveGetUpdate(t *testing.T) {
	db := openTestDB(t)
	truncateData(t, db)

	ctx := context.Background()
	u := newTestUser(t, db, uniqueEmail(t))
	tk := domain.NewTicket(u.ID, "subj", "msg", nil)

	repo := postgres.NewTicketRepository(db)
	if err := repo.Save(ctx, tk); err != nil {
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
	if err := repo.Update(ctx, got); err != nil {
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
		if err := repo.Update(context.Background(), cfg2); err != nil {
			t.Fatalf("cleanup Update: %v", err)
		}
	})

	cfg.Value = domain.NewNumeric(6.71)
	cfg.Description = "SUT"
	if err := repo.Update(ctx, cfg); err != nil {
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
