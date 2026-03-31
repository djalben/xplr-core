package domain

import (
	"strings"
	"time"
)

// TransactionTypeCardSpend — списание с баланса карты (оплата / трата).
const TransactionTypeCardSpend = "CARD_SPEND"

// MaxActiveCardsPerTypeByGrade — максимум активных (не CLOSED) карт одного типа на пользователя.
// STANDARD: 3 каждого вида; GOLD: 5. Иное значение в БД (устаревшее) обрабатывается как STANDARD.
func MaxActiveCardsPerTypeByGrade(grade string) int {
	switch strings.ToUpper(strings.TrimSpace(grade)) {
	case UserGradeGold:
		return 5
	case UserGradeStandard:
		return 3
	default:
		return 3
	}
}

// MonthlySpendLimitByCardType — суммарный лимит трат за календарный месяц (UTC) по всем картам данного типа.
func MonthlySpendLimitByCardType(t CardType) Numeric {
	switch t {
	case CardTypeSubscriptions:
		return NewNumeric(5000)
	case CardTypeTravel:
		return NewNumeric(15000)
	case CardTypePremium:
		return NewNumeric(50000)
	default:
		return NewNumeric(0)
	}
}

// DayBoundsUTC — [start, end) для календарного дня в UTC.
func DayBoundsUTC(now time.Time) (start, end time.Time) {
	y, m, d := now.In(time.UTC).Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	end = start.AddDate(0, 0, 1)

	return start, end
}

// MonthBoundsUTC — [start, end) для календарного месяца в UTC.
func MonthBoundsUTC(now time.Time) (start, end time.Time) {
	y, m, _ := now.In(time.UTC).Date()
	start = time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	end = start.AddDate(0, 1, 0)

	return start, end
}
