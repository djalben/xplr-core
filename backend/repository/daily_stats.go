package repository

import (
	"fmt"
	"log"
)

// DailyStats holds aggregated metrics for the last 24 hours.
type DailyStats struct {
	NewUsers          int     `json:"new_users"`
	NewCards          int     `json:"new_cards"`
	TransactionVolume float64 `json:"transaction_volume"` // sum of successful tx amounts
	TransactionCount  int     `json:"transaction_count"`
}

// GetDailyStats collects platform metrics for the last 24 hours.
func GetDailyStats() (DailyStats, error) {
	if GlobalDB == nil {
		return DailyStats{}, fmt.Errorf("database connection not initialized")
	}

	var stats DailyStats

	// New users in last 24h
	err := GlobalDB.QueryRow(
		`SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '24 hours'`,
	).Scan(&stats.NewUsers)
	if err != nil {
		log.Printf("[DAILY-STATS] Failed to count new users: %v", err)
	}

	// New cards in last 24h
	err = GlobalDB.QueryRow(
		`SELECT COUNT(*) FROM cards WHERE created_at >= NOW() - INTERVAL '24 hours'`,
	).Scan(&stats.NewCards)
	if err != nil {
		log.Printf("[DAILY-STATS] Failed to count new cards: %v", err)
	}

	// Transaction volume and count (successful transactions in last 24h)
	err = GlobalDB.QueryRow(
		`SELECT COALESCE(COUNT(*), 0), COALESCE(SUM(amount), 0)
		 FROM transactions
		 WHERE executed_at >= NOW() - INTERVAL '24 hours'
		   AND status IN ('SUCCESS', 'COMPLETED', 'CAPTURED', 'CAPTURE')`,
	).Scan(&stats.TransactionCount, &stats.TransactionVolume)
	if err != nil {
		log.Printf("[DAILY-STATS] Failed to get transaction volume: %v", err)
	}

	return stats, nil
}
