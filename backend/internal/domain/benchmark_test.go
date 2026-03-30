package domain_test

import (
	"testing"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

// Запуск: go test -bench=. -benchmem ./internal/domain/

func BenchmarkNewNumeric(b *testing.B) {
	for b.Loop() {
		_ = domain.NewNumeric(123.456)
	}
}

func BenchmarkNumeric_Add_Sub(b *testing.B) {
	a := domain.NewNumeric(1000.5)

	c := domain.NewNumeric(0.25)

	for b.Loop() {
		x := a.Add(c)
		_ = x.Sub(c)
	}
}

func BenchmarkNumeric_String(b *testing.B) {
	n := domain.NewNumeric(9999.99)
	for b.Loop() {
		_ = n.String()
	}
}

func BenchmarkParseUUID(b *testing.B) {
	const s = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	for b.Loop() {
		_, err := domain.ParseUUID(s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewUser(b *testing.B) {
	for b.Loop() {
		_, err := domain.NewUser("bench@example.com", "hash")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewWallet(b *testing.B) {
	uid := domain.NewUUID()

	for b.Loop() {
		_ = domain.NewWallet(uid)
	}
}

func BenchmarkWallet_TopUp(b *testing.B) {
	uid := domain.NewUUID()
	amt := domain.NewNumeric(10)
	w := domain.NewWallet(uid)
	for b.Loop() {
		w.Balance = domain.NewNumeric(0)
		err := w.TopUp(amt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWallet_Withdraw(b *testing.B) {
	uid := domain.NewUUID()
	amt := domain.NewNumeric(1)
	w := domain.NewWallet(uid)
	for b.Loop() {
		w.Balance = domain.NewNumeric(1000)
		err := w.Withdraw(amt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewCard(b *testing.B) {
	uid := domain.NewUUID()
	for b.Loop() {
		_, err := domain.NewCard(uid, domain.CardTypeSubscriptions, "prov", "nick")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewTransaction(b *testing.B) {
	uid := domain.NewUUID()

	amount := domain.NewNumeric(50)

	fee := domain.NewNumeric(0)

	for b.Loop() {
		_ = domain.NewTransaction(uid, nil, amount, fee, "TYPE", "OK", "details")
	}
}

func BenchmarkNewTicket(b *testing.B) {
	uid := domain.NewUUID()
	for b.Loop() {
		_ = domain.NewTicket(uid, "subj", "body", nil)
	}
}

func BenchmarkTicket_TakeClose(b *testing.B) {
	uid := domain.NewUUID()

	admin := domain.NewUUID()

	for b.Loop() {
		tk := domain.NewTicket(uid, "s", "m", nil)
		tk.Take(admin)
		tk.Close("reply")
	}
}
