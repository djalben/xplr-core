package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/providers/vless"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/djalben/xplr-core/backend/shop"
	"github.com/djalben/xplr-core/backend/telegram"
)

// VPNTrafficCronHandler is called by Vercel Cron every 30 minutes.
// It fetches Aeza bandwidth + 3X-UI traffic, persists to system_settings,
// and alerts admins if remaining traffic <= 5 GB.
// Protected by CRON_SECRET header check.
func VPNTrafficCronHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	// Verify cron secret
	cronSecret := os.Getenv("CRON_SECRET")
	authHeader := r.Header.Get("Authorization")
	if cronSecret != "" && authHeader != "Bearer "+cronSecret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("[VPN-CRON] 🔄 Starting VPN traffic check...")

	if GlobalDB == nil {
		json.NewEncoder(w).Encode(map[string]any{"error": "DB not initialized"})
		return
	}

	// 1. Fetch bandwidth limit from Aeza API
	limitGB := 30
	aezaInfo, aezaErr := service.GetAezaServerInfo()
	if aezaErr != nil {
		log.Printf("[VPN-CRON] ⚠️ Aeza API error: %v", aezaErr)
		// Try DB-cached value
		var cached string
		if dbErr := GlobalDB.QueryRow(`SELECT setting_value FROM system_settings WHERE setting_key = 'vpn_bandwidth_limit_gb'`).Scan(&cached); dbErr == nil {
			if v, _ := strconv.Atoi(cached); v > 0 {
				limitGB = v
				log.Printf("[VPN-CRON] Using cached DB value: %d GB", limitGB)
			}
		}
	} else if aezaInfo != nil && aezaInfo.BandwidthGB > 0 {
		limitGB = aezaInfo.BandwidthGB
		log.Printf("[VPN-CRON] ✅ Aeza live bandwidth: %d GB", limitGB)
	} else {
		log.Printf("[VPN-CRON] ⚠️ Aeza returned BandwidthGB=0 — parser issue, check [AEZA-SERVER] logs")
	}

	// FORCE WRITE bandwidth to system_settings
	result, dbErr := GlobalDB.Exec(`
		INSERT INTO system_settings (setting_key, setting_value, description)
		VALUES ('vpn_bandwidth_limit_gb', $1, 'VPN server traffic limit from Aeza plan (GB)')
		ON CONFLICT (setting_key) DO UPDATE SET setting_value = $1, updated_at = NOW()`,
		strconv.Itoa(limitGB))
	if dbErr != nil {
		log.Printf("[VPN-CRON] ❌ DB write vpn_bandwidth_limit_gb failed: %v", dbErr)
	} else {
		rows, _ := result.RowsAffected()
		log.Printf("[VPN-CRON] ✅ DB write OK: vpn_bandwidth_limit_gb=%d (rows=%d)", limitGB, rows)
	}

	// 2. Fetch aggregate traffic from 3X-UI panel
	var totalTrafficBytes int64
	provider := shop.GetRegistry().Get("vless")
	if provider != nil {
		if vp, ok := provider.(*vless.VlessProvider); ok {
			stats, err := vp.GetServerTraffic()
			if err != nil {
				log.Printf("[VPN-CRON] ⚠️ 3X-UI traffic error: %v", err)
			} else {
				totalTrafficBytes = stats.TotalTraffic
				log.Printf("[VPN-CRON] 3X-UI total traffic: %.2f GB", float64(totalTrafficBytes)/(1024*1024*1024))
			}
		}
	} else {
		log.Println("[VPN-CRON] ⚠️ VlessProvider not available — skipping traffic fetch")
	}

	// Persist traffic used
	GlobalDB.Exec(`
		INSERT INTO system_settings (setting_key, setting_value, description)
		VALUES ('vpn_traffic_used_bytes', $1, 'VPN server total traffic used (bytes)')
		ON CONFLICT (setting_key) DO UPDATE SET setting_value = $1, updated_at = NOW()`,
		strconv.FormatInt(totalTrafficBytes, 10))

	// Persist last check timestamp
	GlobalDB.Exec(`
		INSERT INTO system_settings (setting_key, setting_value, description)
		VALUES ('vpn_monitor_last_check', NOW()::text, 'Last VPN traffic monitor check timestamp')
		ON CONFLICT (setting_key) DO UPDATE SET setting_value = NOW()::text, updated_at = NOW()`)

	// 3. Check threshold and alert
	limitBytes := int64(limitGB) * 1024 * 1024 * 1024
	remainingBytes := limitBytes - totalTrafficBytes
	if remainingBytes < 0 {
		remainingBytes = 0
	}
	remainingGB := float64(remainingBytes) / (1024 * 1024 * 1024)
	usedGB := float64(totalTrafficBytes) / (1024 * 1024 * 1024)

	log.Printf("[VPN-CRON] ✅ Limit=%dGB, Used=%.2fGB, Remaining=%.2fGB", limitGB, usedGB, remainingGB)

	alertSent := false
	if remainingGB <= 5.0 {
		log.Printf("[VPN-CRON] 🚨 CRITICAL: remaining %.1f GB <= 5 GB — sending alert", remainingGB)
		msg := fmt.Sprintf("🚨 <b>Осталось менее 5 ГБ трафика на VPN-сервере!</b>\n\n"+
			"Остаток: <b>%.1f ГБ</b> из %d ГБ\n"+
			"Использовано: <b>%.1f ГБ</b>\n\n"+
			"⚠️ Рекомендуется увеличить лимит или ограничить новые подключения.",
			remainingGB, limitGB, usedGB)
		telegram.NotifyAdmins(msg, "Открыть админку", "https://xplr.pro/staff-only-zone")
		alertSent = true
	}

	json.NewEncoder(w).Encode(map[string]any{
		"ok":           true,
		"limit_gb":     limitGB,
		"used_gb":      usedGB,
		"remaining_gb": remainingGB,
		"alert_sent":   alertSent,
	})
}

// VPNCleanupCronHandler is called by Vercel Cron every 6 hours.
// 1. Fixes 0/0 GB records in store_orders
// 2. Expires keys where traffic exceeded or time elapsed
// 3. Disables/deletes expired clients from 3X-UI panel
func VPNCleanupCronHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	cronSecret := os.Getenv("CRON_SECRET")
	authHeader := r.Header.Get("Authorization")
	if cronSecret != "" && authHeader != "Bearer "+cronSecret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("[VPN-CLEANUP-CRON] 🔄 Starting cleanup...")

	if GlobalDB == nil {
		json.NewEncoder(w).Encode(map[string]any{"error": "DB not initialized"})
		return
	}

	// ── 1. Fix 0/0 GB records ──
	planQuotas := map[int]int64{
		7:   15 * 1024 * 1024 * 1024,
		30:  60 * 1024 * 1024 * 1024,
		180: 300 * 1024 * 1024 * 1024,
		365: 600 * 1024 * 1024 * 1024,
	}
	var fixedCount int64
	for days, correctBytes := range planQuotas {
		res, err := GlobalDB.Exec(`
			UPDATE store_orders
			SET meta = jsonb_set(COALESCE(meta, '{}'), '{traffic_bytes}', to_jsonb($1::bigint))
			WHERE status = 'completed'
			  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%' OR activation_key LIKE 'vless://%')
			  AND COALESCE((meta->>'duration_days')::int, 0) = $2
			  AND (COALESCE((meta->>'traffic_bytes')::bigint, 0) = 0)`,
			correctBytes, days)
		if err != nil {
			log.Printf("[VPN-CLEANUP-CRON] ❌ Fix 0-byte %d-day: %v", days, err)
		} else if n, _ := res.RowsAffected(); n > 0 {
			fixedCount += n
			log.Printf("[VPN-CLEANUP-CRON] 🔧 Fixed %d orders (%d-day): 0→%dGB", n, days, correctBytes/(1024*1024*1024))
		}
	}

	// Fix orders with no duration_days
	if res, err := GlobalDB.Exec(`
		UPDATE store_orders
		SET meta = jsonb_set(
			jsonb_set(COALESCE(meta, '{}'), '{duration_days}', '30'),
			'{traffic_bytes}', to_jsonb($1::bigint))
		WHERE status = 'completed'
		  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%' OR activation_key LIKE 'vless://%')
		  AND (meta IS NULL OR meta->>'duration_days' IS NULL OR (meta->>'duration_days')::int = 0)
		  AND (meta IS NULL OR meta->>'traffic_bytes' IS NULL OR (meta->>'traffic_bytes')::bigint = 0)`,
		int64(60)*1024*1024*1024); err == nil {
		if n, _ := res.RowsAffected(); n > 0 {
			fixedCount += n
		}
	}

	// ── 2. Expire over-limit and timed-out keys ──
	var vp *vless.VlessProvider
	if p := shop.GetRegistry().Get("vless"); p != nil {
		vp, _ = p.(*vless.VlessProvider)
	}

	var trafficExpired, timeExpired int

	// 2a. Traffic exceeded
	if vp != nil {
		tRows, err := GlobalDB.Query(`
			SELECT id, provider_ref, COALESCE(meta, '{}')
			FROM store_orders
			WHERE status = 'completed' AND provider_ref != ''
			  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%' OR activation_key LIKE 'vless://%')
			  AND COALESCE((meta->>'traffic_bytes')::bigint, 0) > 0`)
		if err == nil {
			defer tRows.Close()
			for tRows.Next() {
				var orderID int
				var ref, metaStr string
				if tRows.Scan(&orderID, &ref, &metaStr) != nil {
					continue
				}
				var meta struct {
					TrafficBytes int64 `json:"traffic_bytes"`
				}
				json.Unmarshal([]byte(metaStr), &meta)
				if meta.TrafficBytes <= 0 {
					continue
				}
				stats, err := vp.GetClientTraffic(ref)
				if err != nil {
					continue
				}
				if stats.Up+stats.Down >= meta.TrafficBytes {
					log.Printf("[VPN-CLEANUP-CRON] 🚫 Traffic exceeded: %s (%d/%d)", ref, stats.Up+stats.Down, meta.TrafficBytes)
					GlobalDB.Exec(`UPDATE store_orders SET status = 'expired' WHERE id = $1`, orderID)
					vp.DeleteClient(ref)
					trafficExpired++
				}
			}
		}
	}

	// 2b. Time expired
	nowMs := time.Now().UnixMilli()
	if tRows, err := GlobalDB.Query(`
		SELECT id, provider_ref
		FROM store_orders
		WHERE status = 'completed' AND provider_ref != ''
		  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%' OR activation_key LIKE 'vless://%')
		  AND COALESCE((meta->>'expire_ms')::bigint, 0) > 0
		  AND (meta->>'expire_ms')::bigint < $1`, nowMs); err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var orderID int
			var ref string
			if tRows.Scan(&orderID, &ref) != nil {
				continue
			}
			log.Printf("[VPN-CLEANUP-CRON] ⏰ Time expired: %s", ref)
			GlobalDB.Exec(`UPDATE store_orders SET status = 'expired' WHERE id = $1`, orderID)
			if vp != nil {
				vp.DeleteClient(ref)
			}
			timeExpired++
		}
	}

	totalExpired := trafficExpired + timeExpired
	if totalExpired > 0 {
		telegram.NotifyAdmins(
			fmt.Sprintf("🧹 <b>VPN Cleanup (cron):</b>\n• Трафик исчерпан: %d\n• Срок истёк: %d\n• Записей исправлено: %d",
				trafficExpired, timeExpired, fixedCount),
			"Открыть админку", "https://xplr.pro/staff-only-zone")
	}

	json.NewEncoder(w).Encode(map[string]any{
		"ok":              true,
		"fixed_records":   fixedCount,
		"traffic_expired": trafficExpired,
		"time_expired":    timeExpired,
	})
}
