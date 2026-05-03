package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/djalben/xplr-core/backend/providers/vless"
	"github.com/djalben/xplr-core/backend/shop"
	"github.com/djalben/xplr-core/backend/telegram"
)

// StartVPNCleanupJob runs every 6 hours:
// 1. Fixes 0/0 GB records — sets default traffic_bytes based on duration_days
// 2. Checks expired keys (time or traffic exceeded) — marks as 'expired', disables in 3X-UI
func StartVPNCleanupJob() {
	go func() {
		// Wait 2 minutes for providers to initialize
		time.Sleep(2 * time.Minute)
		runVPNCleanup()

		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			runVPNCleanup()
		}
	}()
	log.Println("[VPN-CLEANUP] ✅ Auto-cleanup job started (interval=6h)")
}

func runVPNCleanup() {
	if MonitorDB == nil {
		log.Println("[VPN-CLEANUP] ⚠️ DB not initialized — skipping")
		return
	}
	log.Println("[VPN-CLEANUP] 🔄 Running cleanup cycle...")

	fixZeroTrafficRecords()
	expireOverLimitKeys()
	expireTimedOutKeys()

	log.Println("[VPN-CLEANUP] ✅ Cleanup cycle complete")
}

// fixZeroTrafficRecords sets default traffic_bytes for orders with 0 or missing values.
func fixZeroTrafficRecords() {
	planQuotas := map[int]int64{
		7:   15 * 1024 * 1024 * 1024,
		30:  60 * 1024 * 1024 * 1024,
		180: 300 * 1024 * 1024 * 1024,
		365: 600 * 1024 * 1024 * 1024,
	}

	for days, correctBytes := range planQuotas {
		res, err := MonitorDB.Exec(`
			UPDATE store_orders
			SET meta = jsonb_set(COALESCE(meta, '{}'), '{traffic_bytes}', to_jsonb($1::bigint))
			WHERE status = 'completed'
			  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%' OR activation_key LIKE 'vless://%')
			  AND COALESCE((meta->>'duration_days')::int, 0) = $2
			  AND (COALESCE((meta->>'traffic_bytes')::bigint, 0) = 0)`,
			correctBytes, days)
		if err != nil {
			log.Printf("[VPN-CLEANUP] ❌ Failed to fix 0-byte records for %d-day: %v", days, err)
		} else if n, _ := res.RowsAffected(); n > 0 {
			log.Printf("[VPN-CLEANUP] 🔧 Fixed %d orders (%d-day): 0 → %d GB", n, days, correctBytes/(1024*1024*1024))
		}
	}

	// Fix orders with no duration_days at all — default to 30 days / 60 GB
	res, err := MonitorDB.Exec(`
		UPDATE store_orders
		SET meta = jsonb_set(
			jsonb_set(COALESCE(meta, '{}'), '{duration_days}', '30'),
			'{traffic_bytes}', to_jsonb($1::bigint)
		)
		WHERE status = 'completed'
		  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%' OR activation_key LIKE 'vless://%')
		  AND (meta IS NULL OR meta->>'duration_days' IS NULL OR (meta->>'duration_days')::int = 0)
		  AND (meta IS NULL OR meta->>'traffic_bytes' IS NULL OR (meta->>'traffic_bytes')::bigint = 0)`,
		int64(60)*1024*1024*1024)
	if err != nil {
		log.Printf("[VPN-CLEANUP] ❌ Failed to fix null-duration records: %v", err)
	} else if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("[VPN-CLEANUP] 🔧 Fixed %d orders with no duration → 30d/60GB default", n)
	}
}

// expireOverLimitKeys finds completed VPN orders where traffic is exceeded, marks as expired, disables in panel.
func expireOverLimitKeys() {
	rows, err := MonitorDB.Query(`
		SELECT id, provider_ref, COALESCE(meta, '{}'), COALESCE(activation_key, '')
		FROM store_orders
		WHERE status = 'completed'
		  AND provider_ref != ''
		  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%' OR activation_key LIKE 'vless://%')
		  AND COALESCE((meta->>'traffic_bytes')::bigint, 0) > 0`)
	if err != nil {
		log.Printf("[VPN-CLEANUP] ❌ Query expired-traffic error: %v", err)
		return
	}
	defer rows.Close()

	vp := getCleanupVlessProvider()

	var expiredCount int
	for rows.Next() {
		var orderID int
		var ref, metaStr, activationKey string
		if err := rows.Scan(&orderID, &ref, &metaStr, &activationKey); err != nil {
			continue
		}

		var meta struct {
			TrafficBytes int64 `json:"traffic_bytes"`
			ExpireMs     int64 `json:"expire_ms"`
		}
		json.Unmarshal([]byte(metaStr), &meta)

		if meta.TrafficBytes <= 0 {
			continue
		}

		// Query live traffic from panel
		if vp == nil {
			continue
		}
		stats, err := vp.GetClientTraffic(ref)
		if err != nil {
			continue // Client might not exist on panel
		}

		usedBytes := stats.Up + stats.Down
		if usedBytes < meta.TrafficBytes {
			continue // Still within limit
		}

		// Traffic exceeded — expire
		log.Printf("[VPN-CLEANUP] 🚫 Traffic exceeded for %s: %d/%d bytes — expiring", ref, usedBytes, meta.TrafficBytes)
		expireOrder(orderID, ref, vp, "traffic_exceeded")
		expiredCount++
	}

	if expiredCount > 0 {
		log.Printf("[VPN-CLEANUP] 🚫 Expired %d orders (traffic exceeded)", expiredCount)
		telegram.NotifyAdmins(
			fmt.Sprintf("🧹 <b>VPN Cleanup:</b> %d ключей деактивировано (трафик исчерпан)", expiredCount),
			"Открыть админку", "https://xplr.pro/staff-only-zone")
	}
}

// expireTimedOutKeys finds completed VPN orders past their expiry time.
func expireTimedOutKeys() {
	nowMs := time.Now().UnixMilli()

	rows, err := MonitorDB.Query(`
		SELECT id, provider_ref
		FROM store_orders
		WHERE status = 'completed'
		  AND provider_ref != ''
		  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%' OR activation_key LIKE 'vless://%')
		  AND COALESCE((meta->>'expire_ms')::bigint, 0) > 0
		  AND (meta->>'expire_ms')::bigint < $1`, nowMs)
	if err != nil {
		log.Printf("[VPN-CLEANUP] ❌ Query expired-time error: %v", err)
		return
	}
	defer rows.Close()

	vp := getCleanupVlessProvider()

	var expiredCount int
	for rows.Next() {
		var orderID int
		var ref string
		if err := rows.Scan(&orderID, &ref); err != nil {
			continue
		}

		log.Printf("[VPN-CLEANUP] ⏰ Key expired (time) for %s — deactivating", ref)
		expireOrder(orderID, ref, vp, "time_expired")
		expiredCount++
	}

	if expiredCount > 0 {
		log.Printf("[VPN-CLEANUP] ⏰ Expired %d orders (time elapsed)", expiredCount)
		telegram.NotifyAdmins(
			fmt.Sprintf("🧹 <b>VPN Cleanup:</b> %d ключей деактивировано (срок истёк)", expiredCount),
			"Открыть админку", "https://xplr.pro/staff-only-zone")
	}
}

// expireOrder marks order as 'expired' in DB and disables/deletes client on 3X-UI panel.
func expireOrder(orderID int, ref string, vp *vless.VlessProvider, reason string) {
	// Update DB status
	_, err := MonitorDB.Exec(`UPDATE store_orders SET status = 'expired' WHERE id = $1`, orderID)
	if err != nil {
		log.Printf("[VPN-CLEANUP] ❌ Failed to update order %d status: %v", orderID, err)
	}

	// Try to disable on panel
	if vp != nil {
		if err := vp.DeleteClient(ref); err != nil {
			log.Printf("[VPN-CLEANUP] ⚠️ Failed to delete client %s from panel: %v (reason: %s)", ref, err, reason)
		} else {
			log.Printf("[VPN-CLEANUP] ✅ Deleted client %s from panel (reason: %s)", ref, reason)
		}
	}
}

func getCleanupVlessProvider() *vless.VlessProvider {
	provider := shop.GetRegistry().Get("vless")
	if provider == nil {
		return nil
	}
	vp, _ := provider.(*vless.VlessProvider)
	return vp
}
