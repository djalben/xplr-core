package service

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/providers/vless"
	"github.com/djalben/xplr-core/backend/shop"
	"github.com/djalben/xplr-core/backend/telegram"
)

// MonitorDB is set by main.go to the global *sql.DB so the monitor can persist data.
var MonitorDB *sql.DB

// getVlessProvider resolves the VlessProvider from the shop registry at runtime.
func getVlessProvider() *vless.VlessProvider {
	provider := shop.GetRegistry().Get("vless")
	if provider == nil {
		return nil
	}
	vp, _ := provider.(*vless.VlessProvider)
	return vp
}

// StartVPNTrafficMonitor starts a background goroutine that every 30 minutes:
// 1. Fetches the current bandwidth limit from Aeza API
// 2. Fetches aggregate traffic from 3X-UI panel
// 3. Persists both to system_settings
// 4. Sends a Telegram alert if remaining traffic <= 5 GB
func StartVPNTrafficMonitor() {
	apiKey := os.Getenv("AEZA_API_KEY")
	if apiKey == "" {
		log.Println("[VPN-MONITOR] ⚠️ AEZA_API_KEY not set — VPN traffic monitor disabled")
		return
	}

	go func() {
		// Initial check after 60 seconds (let providers initialize)
		time.Sleep(60 * time.Second)
		runVPNTrafficCheck()

		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			runVPNTrafficCheck()
		}
	}()

	log.Println("[VPN-MONITOR] ✅ VPN traffic monitor started (interval=30m, threshold=5GB)")
}

func runVPNTrafficCheck() {
	log.Println("[VPN-MONITOR] 🔄 Running traffic check...")

	// 1. Fetch bandwidth limit from Aeza API
	limitGB := 30 // default fallback
	aezaInfo, err := GetAezaServerInfo()
	if err != nil {
		log.Printf("[VPN-MONITOR] ⚠️ Aeza API error: %v — using cached/default limit", err)
		// Try to read cached value from DB
		if MonitorDB != nil {
			var cached string
			if dbErr := MonitorDB.QueryRow(
				`SELECT setting_value FROM system_settings WHERE setting_key = 'vpn_bandwidth_limit_gb'`,
			).Scan(&cached); dbErr == nil {
				if v, _ := strconv.Atoi(cached); v > 0 {
					limitGB = v
				}
			}
		}
	} else if aezaInfo.BandwidthGB > 0 {
		limitGB = aezaInfo.BandwidthGB
		log.Printf("[VPN-MONITOR] Aeza bandwidth: %d GB", limitGB)
	}

	// Persist bandwidth limit to system_settings
	if MonitorDB != nil {
		upsertSetting("vpn_bandwidth_limit_gb", strconv.Itoa(limitGB), "VPN server traffic limit from Aeza plan (GB)")
	}

	// 2. Fetch aggregate traffic from 3X-UI panel
	var totalTrafficBytes int64
	vp := getVlessProvider()
	if vp != nil {
		stats, err := vp.GetServerTraffic()
		if err != nil {
			log.Printf("[VPN-MONITOR] ⚠️ 3X-UI traffic error: %v", err)
		} else {
			totalTrafficBytes = stats.TotalTraffic
			log.Printf("[VPN-MONITOR] 3X-UI total traffic: %.2f GB (%d bytes)",
				float64(totalTrafficBytes)/(1024*1024*1024), totalTrafficBytes)
		}
	} else {
		log.Println("[VPN-MONITOR] ⚠️ VlessProvider not available — skipping traffic fetch")
		// Try env fallback for limit
		if envLimit := os.Getenv("VPN_SERVER_TRAFFIC_LIMIT_GB"); envLimit != "" {
			if v, _ := strconv.Atoi(envLimit); v > 0 {
				limitGB = v
			}
		}
	}

	// Persist used traffic to system_settings
	if MonitorDB != nil {
		upsertSetting("vpn_traffic_used_bytes", strconv.FormatInt(totalTrafficBytes, 10), "VPN server total traffic used (bytes)")
		upsertSetting("vpn_monitor_last_check", time.Now().Format(time.RFC3339), "Last VPN traffic monitor check timestamp")
	}

	// 3. Check threshold and alert
	limitBytes := int64(limitGB) * 1024 * 1024 * 1024
	remainingBytes := limitBytes - totalTrafficBytes
	if remainingBytes < 0 {
		remainingBytes = 0
	}
	remainingGB := float64(remainingBytes) / (1024 * 1024 * 1024)

	log.Printf("[VPN-MONITOR] ✅ Limit=%dGB, Used=%.2fGB, Remaining=%.2fGB",
		limitGB, float64(totalTrafficBytes)/(1024*1024*1024), remainingGB)

	if remainingGB <= 5.0 {
		log.Printf("[VPN-MONITOR] 🚨 CRITICAL: remaining %.1f GB <= 5 GB — sending alert", remainingGB)
		sendTrafficAlert(totalTrafficBytes, limitBytes)
	}
}

func sendTrafficAlert(totalTraffic, limitBytes int64) {
	usedGB := float64(totalTraffic) / (1024 * 1024 * 1024)
	limitGB := float64(limitBytes) / (1024 * 1024 * 1024)
	remainingGB := limitGB - usedGB
	if remainingGB < 0 {
		remainingGB = 0
	}
	msg := fmt.Sprintf("🚨 <b>Осталось менее 5 ГБ трафика на VPN-сервере!</b>\n\n"+
		"Остаток: <b>%.1f ГБ</b> из %.0f ГБ\n"+
		"Использовано: <b>%.1f ГБ</b>\n\n"+
		"⚠️ Рекомендуется увеличить лимит или ограничить новые подключения.",
		remainingGB, limitGB, usedGB)
	telegram.NotifyAdmins(msg, "Открыть админку", "https://xplr.pro/staff-only-zone")
}

func upsertSetting(key, value, description string) {
	if MonitorDB == nil {
		return
	}
	_, err := MonitorDB.Exec(`
		INSERT INTO system_settings (setting_key, setting_value, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (setting_key) DO UPDATE SET setting_value = $2, updated_at = NOW()
	`, key, value, description)
	if err != nil {
		log.Printf("[VPN-MONITOR] ⚠️ Failed to upsert setting %s: %v", key, err)
	}
}
