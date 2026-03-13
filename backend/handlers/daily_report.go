package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/telegram"
)

// SendDailyReportHandler — GET /api/v1/admin/send-daily-report
// Protected by a secret key (query param or header) for internal/cron use.
// Usage: GET /api/v1/admin/send-daily-report?key=YOUR_SECRET
func SendDailyReportHandler(w http.ResponseWriter, r *http.Request) {
	// Verify secret key
	secret := os.Getenv("DAILY_REPORT_SECRET")
	if secret == "" {
		secret = "xplr-daily-report-2024" // fallback default
	}

	providedKey := r.URL.Query().Get("key")
	if providedKey == "" {
		providedKey = r.Header.Get("X-Report-Key")
	}
	if providedKey != secret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Collect stats
	stats, err := repository.GetDailyStats()
	if err != nil {
		log.Printf("[DAILY-REPORT] Failed to get stats: %v", err)
		http.Error(w, "Failed to collect stats", http.StatusInternalServerError)
		return
	}

	// Format the report
	now := time.Now().Format("02.01.2006")
	msg := fmt.Sprintf(
		"📊 <b>Ежедневный отчёт XPLR</b>\n"+
			"📅 %s\n\n"+
			"👤 <b>Новые пользователи:</b> %d\n"+
			"💳 <b>Новые карты:</b> %d\n"+
			"🔄 <b>Транзакций:</b> %d\n"+
			"💰 <b>Оборот:</b> $%.2f\n\n"+
			"📈 Данные за последние 24 часа.",
		now,
		stats.NewUsers,
		stats.NewCards,
		stats.TransactionCount,
		stats.TransactionVolume,
	)

	// Send to all admins
	telegram.NotifyAdmins(msg, "📊 Открыть дашборд", "https://xplr.pro/admin/dashboard")

	log.Printf("[DAILY-REPORT] ✅ Report sent: users=%d, cards=%d, txs=%d, volume=%.2f",
		stats.NewUsers, stats.NewCards, stats.TransactionCount, stats.TransactionVolume)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "sent",
		"stats":  stats,
	})
}
