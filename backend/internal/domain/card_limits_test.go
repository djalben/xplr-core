package domain_test

import (
	"testing"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

func TestMaxActiveCardsPerTypeByGrade(t *testing.T) {
	t.Parallel()

	tests := []struct {
		grade string
		want  int
	}{
		{"STANDARD", 3},
		{"standard", 3},
		{" GOLD ", 5},
		{"GOLD", 5},
		{"SILVER", 3},
		{"PLATINUM", 3},
		{"", 3},
	}
	for _, tt := range tests {
		t.Run(tt.grade, func(t *testing.T) {
			t.Parallel()
			if got := domain.MaxActiveCardsPerTypeByGrade(tt.grade); got != tt.want {
				t.Fatalf("MaxActiveCardsPerTypeByGrade(%q) = %d, want %d", tt.grade, got, tt.want)
			}
		})
	}
}

func TestMonthlySpendLimitByCardType(t *testing.T) {
	t.Parallel()

	if domain.MonthlySpendLimitByCardType(domain.CardTypeSubscriptions).Equal(domain.NewNumeric(5000)) != true {
		t.Fatal("subscriptions cap")
	}
	if domain.MonthlySpendLimitByCardType(domain.CardTypeTravel).Equal(domain.NewNumeric(15000)) != true {
		t.Fatal("travel cap")
	}
	if domain.MonthlySpendLimitByCardType(domain.CardTypePremium).Equal(domain.NewNumeric(50000)) != true {
		t.Fatal("premium cap")
	}
}

func TestDayMonthBoundsUTC(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC)
	ds, de := domain.DayBoundsUTC(now)
	if !ds.Equal(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)) || !de.Equal(time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("day bounds: %v %v", ds, de)
	}
	ms, me := domain.MonthBoundsUTC(now)
	if !ms.Equal(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)) || !me.Equal(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("month bounds: %v %v", ms, me)
	}
}
