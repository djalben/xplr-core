package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// fallbackRubUsdRate is the hardcoded RUB→USD rate used when the DB rate is unavailable.
var fallbackRubUsdRate = decimal.NewFromFloat(89.45)

// DashboardStatsResult holds all data for the user dashboard.
type DashboardStatsResult struct {
	TodayTotal         string              `json:"today_total"`
	TodayCount         int                 `json:"today_count"`
	RecentTransactions []DashboardTx       `json:"recent_transactions"`
	WeeklyChart        []DashboardChartDay `json:"weekly_chart"`
}

// DashboardTx is a simplified transaction for the dashboard recent-ops list.
type DashboardTx struct {
	ID              int       `json:"id"`
	TypeLabel       string    `json:"type_label"`
	TransactionType string    `json:"transaction_type"`
	Amount          string    `json:"amount"`
	Currency        string    `json:"currency"`
	CardLast4       string    `json:"card_last4"`
	Details         string    `json:"details"`
	ExecutedAt      time.Time `json:"executed_at"`
}

// DashboardChartDay represents one day in the 7-day spending chart.
type DashboardChartDay struct {
	Date   string `json:"date"`
	Label  string `json:"label"`
	Amount string `json:"amount"`
}

// txTypeLabel maps transaction_type to a human-readable Russian label.
func txTypeLabel(txType string) string {
	switch txType {
	case "WALLET_TOPUP":
		return "Пополнение"
	case "CARD_ISSUE_FEE":
		return "Выпуск карты"
	case "CARD_TOPUP":
		return "Перевод на карту"
	case "CARD_REFUND":
		return "Возврат"
	case "WALLET_RECLAIM":
		return "Возврат (истёкшая)"
	case "CARD_CHARGE":
		return "Списание"
	case "REFERRAL_BONUS":
		return "Реферальный бонус"
	case "COMMISSION":
		return "Комиссия"
	default:
		return "Операция"
	}
}

// GetDashboardStats returns aggregated dashboard data for a user:
// - today_total: sum of absolute amounts for today (in USD)
// - today_count: number of transactions today
// - recent_transactions: 5 most recent transactions
// - weekly_chart: expenses grouped by day for last 7 days
func GetDashboardStats(userID int) (*DashboardStatsResult, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	result := &DashboardStatsResult{}

	// Expense-only transaction types (excludes top-ups, refunds, bonuses)
	expenseFilter := `AND transaction_type NOT IN ('WALLET_TOPUP', 'CARD_REFUND', 'WALLET_RECLAIM', 'REFERRAL_BONUS')`
	successFilter := `AND status IN ('SUCCESS', 'COMPLETED', 'CAPTURED', 'CAPTURE', 'APPROVED')`

	// ── 1a. Today count (all transactions, informational) ──
	var todayCount int
	_ = GlobalDB.QueryRow(`
		SELECT COUNT(*) FROM transactions
		WHERE user_id = $1 AND executed_at >= CURRENT_DATE `+successFilter,
		userID).Scan(&todayCount)

	// ── 1b. Today spending total (expenses only, in USD) ──
	var todaySum decimal.Decimal
	err := GlobalDB.QueryRow(`
		SELECT COALESCE(SUM(ABS(amount)), 0)
		FROM transactions
		WHERE user_id = $1
		  AND executed_at >= CURRENT_DATE
		  `+successFilter+`
		  `+expenseFilter,
		userID).Scan(&todaySum)
	if err != nil {
		log.Printf("[DASHBOARD-STATS] Error fetching today total for user %d: %v", userID, err)
		todaySum = decimal.Zero
	}

	// Convert RUB expense transactions to USD if needed
	var todayRubSum decimal.Decimal
	err = GlobalDB.QueryRow(`
		SELECT COALESCE(SUM(ABS(amount)), 0)
		FROM transactions
		WHERE user_id = $1
		  AND executed_at >= CURRENT_DATE
		  `+successFilter+`
		  `+expenseFilter+`
		  AND UPPER(COALESCE(currency, 'USD')) = 'RUB'
	`, userID).Scan(&todayRubSum)
	if err == nil && todayRubSum.GreaterThan(decimal.Zero) {
		rate, rateErr := GetFinalRate("RUB", "USD")
		if rateErr != nil || !rate.GreaterThan(decimal.Zero) {
			rate = fallbackRubUsdRate
		}
		rubToUsd := todayRubSum.Div(rate).Round(2)
		todaySum = todaySum.Sub(todayRubSum).Add(rubToUsd)
	}
	result.TodayTotal = todaySum.StringFixed(2)
	result.TodayCount = todayCount

	// ── 2. Recent 5 transactions ──
	rows, err := GlobalDB.Query(`
		SELECT t.id, t.transaction_type, t.amount, COALESCE(t.currency, 'USD'),
		       COALESCE(t.details, ''), t.executed_at,
		       COALESCE(c.last_4_digits, '')
		FROM transactions t
		LEFT JOIN cards c ON t.card_id = c.id
		WHERE t.user_id = $1
		ORDER BY t.executed_at DESC
		LIMIT 5
	`, userID)
	if err != nil {
		log.Printf("[DASHBOARD-STATS] Error fetching recent txns for user %d: %v", userID, err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var tx DashboardTx
			var amt decimal.Decimal
			var execAt time.Time
			err := rows.Scan(&tx.ID, &tx.TransactionType, &amt, &tx.Currency, &tx.Details, &execAt, &tx.CardLast4)
			if err != nil {
				log.Printf("[DASHBOARD-STATS] Error scanning tx: %v", err)
				continue
			}
			tx.ExecutedAt = execAt
			tx.TypeLabel = txTypeLabel(tx.TransactionType)
			tx.Amount = amt.StringFixed(2)
			result.RecentTransactions = append(result.RecentTransactions, tx)
		}
	}
	if result.RecentTransactions == nil {
		result.RecentTransactions = []DashboardTx{}
	}

	// ── 3. Weekly chart: expenses per day for last 7 days ──
	dayLabels := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}

	// Build 7-day scaffold
	now := time.Now()
	type dayEntry struct {
		date  string
		label string
		sum   decimal.Decimal
	}
	days := make([]dayEntry, 7)
	for i := 0; i < 7; i++ {
		d := now.AddDate(0, 0, i-6)
		days[i] = dayEntry{
			date:  d.Format("2006-01-02"),
			label: dayLabels[d.Weekday()],
			sum:   decimal.Zero,
		}
	}

	// Query grouped expenses (spending only — exclude top-ups, refunds, bonuses)
	weekStart := now.AddDate(0, 0, -6).Format("2006-01-02")
	weekRows, err := GlobalDB.Query(`
		SELECT DATE(executed_at) AS day, COALESCE(currency, 'USD') AS curr, SUM(ABS(amount)) AS total
		FROM transactions
		WHERE user_id = $1
		  AND executed_at >= $2::date
		  AND status IN ('SUCCESS', 'COMPLETED', 'CAPTURED', 'CAPTURE', 'APPROVED')
		  AND transaction_type NOT IN ('WALLET_TOPUP', 'CARD_REFUND', 'WALLET_RECLAIM', 'REFERRAL_BONUS')
		GROUP BY day, curr
		ORDER BY day
	`, userID, weekStart)
	if err != nil {
		log.Printf("[DASHBOARD-STATS] Error fetching weekly chart for user %d: %v", userID, err)
	} else {
		defer weekRows.Close()
		// Preload RUB→USD rate once (fallback to 89.45)
		rubRate, rateErr := GetFinalRate("RUB", "USD")
		if rateErr != nil || !rubRate.GreaterThan(decimal.Zero) {
			rubRate = fallbackRubUsdRate
		}

		for weekRows.Next() {
			var dayDate time.Time
			var curr string
			var total decimal.Decimal
			if err := weekRows.Scan(&dayDate, &curr, &total); err != nil {
				continue
			}
			dateStr := dayDate.Format("2006-01-02")
			// Convert to USD if RUB
			amtUsd := total
			if curr == "RUB" && rubRate.GreaterThan(decimal.Zero) {
				amtUsd = total.Div(rubRate).Round(2)
			}
			for j := range days {
				if days[j].date == dateStr {
					days[j].sum = days[j].sum.Add(amtUsd)
					break
				}
			}
		}
	}

	result.WeeklyChart = make([]DashboardChartDay, 7)
	for i, d := range days {
		result.WeeklyChart[i] = DashboardChartDay{
			Date:   d.date,
			Label:  d.label,
			Amount: d.sum.StringFixed(2),
		}
	}

	return result, nil
}

// GetDashboardTodayTotal returns just the today spending total (used for quick refresh).
func GetDashboardTodayTotal(userID int) (string, int, error) {
	if GlobalDB == nil {
		return "0.00", 0, fmt.Errorf("database connection not initialized")
	}
	var sum decimal.Decimal
	var count int
	err := GlobalDB.QueryRow(`
		SELECT COALESCE(SUM(ABS(amount)), 0), COUNT(*)
		FROM transactions
		WHERE user_id = $1
		  AND executed_at >= CURRENT_DATE
		  AND status IN ('SUCCESS', 'COMPLETED', 'CAPTURED', 'CAPTURE', 'APPROVED')
	`, userID).Scan(&sum, &count)
	if err != nil && err != sql.ErrNoRows {
		return "0.00", 0, err
	}
	return sum.StringFixed(2), count, nil
}
