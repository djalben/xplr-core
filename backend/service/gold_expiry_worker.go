package service

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// GoldExpiryWorker checks for users whose Gold tier is about to expire
// and sends notifications at specific day thresholds: 30, 15, 10, 5, 3, 1.
// Should be called periodically (e.g. once per day via cron or ticker).

var goldExpiryDB *sql.DB

func InitGoldExpiryWorker(db *sql.DB) {
	goldExpiryDB = db
}

// RunGoldExpiryCheck finds Gold users expiring in exactly N days and notifies them.
func RunGoldExpiryCheck() {
	if goldExpiryDB == nil {
		log.Println("[GOLD-EXPIRY] ❌ DB not initialized, skipping")
		return
	}

	thresholds := []int{30, 15, 10, 5, 3, 1}

	for _, days := range thresholds {
		notifyUsersExpiringIn(days)
	}

	// Also downgrade users whose Gold has already expired
	downgradeExpiredGoldUsers()
}

func notifyUsersExpiringIn(days int) {
	// Find users whose tier_expires_at falls within today + days (±12 hours window)
	target := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	windowStart := target.Add(-12 * time.Hour)
	windowEnd := target.Add(12 * time.Hour)

	rows, err := goldExpiryDB.Query(`
		SELECT id FROM users 
		WHERE tier = 'gold' 
		  AND tier_expires_at BETWEEN $1 AND $2
	`, windowStart, windowEnd)
	if err != nil {
		log.Printf("[GOLD-EXPIRY] ❌ Error querying %d-day threshold: %v", days, err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			continue
		}

		daysWord := formatDaysRu(days)
		msg := fmt.Sprintf(
			"⏳ <b>Gold заканчивается через %d %s</b>\n\n"+
				"Продлите заранее, чтобы сохранить выгодные условия и лимит 15 карт!\n\n"+
				"<a href=\"https://xplr.pro/dashboard\">Продлить Gold</a>",
			days, daysWord,
		)

		go NotifyUser(userID, "Gold истекает", msg)
		count++
	}

	if count > 0 {
		log.Printf("[GOLD-EXPIRY] 📩 Sent %d-day expiry notifications to %d users", days, count)
	}
}

func downgradeExpiredGoldUsers() {
	res, err := goldExpiryDB.Exec(`
		UPDATE users SET tier = 'standard' 
		WHERE tier = 'gold' AND tier_expires_at < NOW()
	`)
	if err != nil {
		log.Printf("[GOLD-EXPIRY] ❌ Error downgrading expired users: %v", err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("[GOLD-EXPIRY] ⬇️ Downgraded %d expired Gold users to Standard", n)
	}
}

// StartGoldExpiryTicker runs the check once per day in background.
func StartGoldExpiryTicker(db *sql.DB) {
	InitGoldExpiryWorker(db)

	// Run immediately on startup
	go func() {
		time.Sleep(10 * time.Second) // wait for app to fully init
		log.Println("[GOLD-EXPIRY] 🔄 Running initial Gold expiry check...")
		RunGoldExpiryCheck()
	}()

	// Then run every 24 hours
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("[GOLD-EXPIRY] 🔄 Running daily Gold expiry check...")
			RunGoldExpiryCheck()
		}
	}()

	log.Println("[GOLD-EXPIRY] ✅ Gold expiry ticker started (24h interval)")
}

// formatDaysRu returns the correct Russian plural form for "день/дня/дней".
func formatDaysRu(n int) string {
	abs := n
	if abs < 0 {
		abs = -abs
	}
	mod10 := abs % 10
	mod100 := abs % 100
	if mod10 == 1 && mod100 != 11 {
		return "день"
	}
	if mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20) {
		return "дня"
	}
	return "дней"
}
