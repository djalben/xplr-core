package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/infrastructure/persistence/postgres"
	"github.com/jmoiron/sqlx"
)

// Интеграционные SUT-тесты: реальный Postgres. Без DSN тесты пропускаются.
// Пример: XPLR_TEST_POSTGRES_DSN='postgresql://xplr_user:pass@localhost:5432/xplr_db?sslmode=disable'

func testDSN(t *testing.T) string {
	t.Helper()

	dsn := os.Getenv("XPLR_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("задайте XPLR_TEST_POSTGRES_DSN для SUT-тестов репозиториев")
	}

	return dsn
}

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	ctx := context.Background()
	db, err := postgres.Connect(ctx, testDSN(t))
	if err != nil {
		t.Fatalf("postgres connect: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func truncateData(t *testing.T, db *sqlx.DB) {
	t.Helper()

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		TRUNCATE TABLE referrals, tickets, transactions, cards, wallets, user_grades, api_keys, users RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func uniqueEmail(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("%s@example.test", domain.NewUUID().String())
}

func newTestUser(t *testing.T, db *sqlx.DB, email string) *domain.User {
	t.Helper()

	u, err := domain.NewUser(email, "sut-password-hash")
	if err != nil {
		t.Fatalf("domain.NewUser: %v", err)
	}

	repo := postgres.NewUserRepository(db)
	if err := repo.Save(context.Background(), u); err != nil {
		t.Fatalf("user save: %v", err)
	}

	return u
}
