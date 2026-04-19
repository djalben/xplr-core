package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/providers/vless"
	"github.com/djalben/xplr-core/backend/shop"
	"github.com/djalben/xplr-core/backend/telegram"
)

// ══════════════════════════════════════════════════════════════
// GET /api/v1/sub/{ref} — Public VPN subscription endpoint.
// Returns vless:// config as base64 body + Subscription-Userinfo header.
// Called by v2rayNG / Happ Proxy to display traffic progress bar.
// ══════════════════════════════════════════════════════════════

func VPNSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	// Extract {ref} from URL path: /api/v1/sub/{ref}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	var ref string
	for i, p := range parts {
		if p == "sub" && i+1 < len(parts) {
			ref = parts[i+1]
			break
		}
	}
	if ref == "" {
		http.Error(w, "missing subscription ref", http.StatusBadRequest)
		return
	}

	if GlobalDB == nil {
		http.Error(w, "DB not ready", http.StatusInternalServerError)
		return
	}

	// Look up the order by provider_ref (client email tag)
	var activationKey string
	var metaStr string
	err := GlobalDB.QueryRow(`
		SELECT COALESCE(activation_key, ''), COALESCE(meta, '{}')
		FROM store_orders
		WHERE provider_ref = $1 AND status = 'completed'
		ORDER BY created_at DESC LIMIT 1
	`, ref).Scan(&activationKey, &metaStr)
	if err != nil || activationKey == "" {
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}

	// Parse order meta for traffic_bytes and expire_ms
	var meta struct {
		TrafficBytes int64 `json:"traffic_bytes"`
		ExpireMs     int64 `json:"expire_ms"`
	}
	json.Unmarshal([]byte(metaStr), &meta)

	// Query live traffic from 3X-UI panel
	var upload, download int64
	provider := shop.GetRegistry().Get("vless")
	if provider != nil {
		if vp, ok := provider.(*vless.VlessProvider); ok {
			if stats, err := vp.GetClientTraffic(ref); err == nil {
				upload = stats.Up
				download = stats.Down
			}
		}
	}

	// Set Subscription-Userinfo header (standard for v2rayNG / Happ / Clash)
	// Format: upload=X; download=Y; total=Z; expire=T
	total := meta.TrafficBytes
	expireSec := meta.ExpireMs / 1000
	w.Header().Set("Subscription-Userinfo",
		fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d",
			upload, download, total, expireSec))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"XPLR-%s\"", ref))

	// Body: base64-encoded vless:// link (standard subscription format)
	encoded := base64.StdEncoding.EncodeToString([]byte(activationKey))
	w.Write([]byte(encoded))
}

// ══════════════════════════════════════════════════════════════
// GET /api/v1/user/store/vpn-status?ref={email} — VPN key traffic status
// Returns upload/download/total/remaining for the user's VPN key.
// ══════════════════════════════════════════════════════════════

func VPNKeyStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ref := r.URL.Query().Get("ref")
	if ref == "" {
		http.Error(w, "ref parameter required", http.StatusBadRequest)
		return
	}

	// Verify this order belongs to this user
	var metaStr string
	err := GlobalDB.QueryRow(`
		SELECT COALESCE(meta, '{}')
		FROM store_orders
		WHERE provider_ref = $1 AND user_id = $2 AND status = 'completed'
		ORDER BY created_at DESC LIMIT 1
	`, ref, userID).Scan(&metaStr)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	var meta struct {
		TrafficBytes int64 `json:"traffic_bytes"`
		ExpireMs     int64 `json:"expire_ms"`
		DurationDays int   `json:"duration_days"`
	}
	json.Unmarshal([]byte(metaStr), &meta)

	// Query live traffic from panel
	var upload, download int64
	var enabled bool = true
	provider := shop.GetRegistry().Get("vless")
	if provider != nil {
		if vp, ok := provider.(*vless.VlessProvider); ok {
			if stats, err := vp.GetClientTraffic(ref); err == nil {
				upload = stats.Up
				download = stats.Down
				enabled = stats.Enable
			}
		}
	}

	used := upload + download
	total := meta.TrafficBytes
	remaining := total - used
	if remaining < 0 {
		remaining = 0
	}
	exhausted := remaining == 0 && total > 0

	// Determine status
	status := "active"
	if !enabled {
		status = "disabled"
	}
	if exhausted {
		status = "traffic_exhausted"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ref":           ref,
		"status":        status,
		"upload":        upload,
		"download":      download,
		"used":          used,
		"total":         total,
		"remaining":     remaining,
		"exhausted":     exhausted,
		"expire_ms":     meta.ExpireMs,
		"duration_days": meta.DurationDays,
		"used_percent":  safePercent(used, total),
	})
}

func safePercent(used, total int64) float64 {
	if total <= 0 {
		return 0
	}
	pct := float64(used) / float64(total) * 100
	if pct > 100 {
		return 100
	}
	return pct
}

// ══════════════════════════════════════════════════════════════
// GET /api/v1/admin/infra/vpn-server-status — Admin server monitoring
// Returns aggregate traffic, active clients, % of server limit,
// and financial metrics (revenue, cost, margin).
// ══════════════════════════════════════════════════════════════

func AdminVPNServerStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	provider := shop.GetRegistry().Get("vless")
	if provider == nil {
		json.NewEncoder(w).Encode(map[string]any{"error": "vless provider not registered"})
		return
	}

	vp, ok := provider.(*vless.VlessProvider)
	if !ok {
		json.NewEncoder(w).Encode(map[string]any{"error": "provider type assertion failed"})
		return
	}

	stats, err := vp.GetServerTraffic()
	if err != nil {
		log.Printf("[VPN-SERVER-STATUS] ❌ GetServerTraffic error: %v", err)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	// Server traffic limit from env (in GB, default 1000 GB = 1 TB)
	serverLimitGB := 1000
	if envLimit := os.Getenv("VPN_SERVER_TRAFFIC_LIMIT_GB"); envLimit != "" {
		if v, err := strconv.Atoi(envLimit); err == nil && v > 0 {
			serverLimitGB = v
		}
	}
	serverLimitBytes := int64(serverLimitGB) * 1024 * 1024 * 1024

	usedPercent := float64(0)
	if serverLimitBytes > 0 {
		usedPercent = float64(stats.TotalTraffic) / float64(serverLimitBytes) * 100
		if usedPercent > 100 {
			usedPercent = 100
		}
	}

	// ── Financial metrics ──
	// Monthly server cost from env (EUR)
	serverCostEUR := 4.94
	if envCost := os.Getenv("VPN_SERVER_MONTHLY_COST"); envCost != "" {
		if v, err := strconv.ParseFloat(envCost, 64); err == nil && v > 0 {
			serverCostEUR = v
		}
	}

	// Revenue: sum of all completed VPN orders this month
	var monthlyRevenue float64
	if GlobalDB != nil {
		_ = GlobalDB.QueryRow(`
			SELECT COALESCE(SUM(price_usd), 0)
			FROM store_orders
			WHERE status = 'completed'
			  AND product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%'
			  AND created_at >= date_trunc('month', NOW())
		`).Scan(&monthlyRevenue)
	}

	margin := monthlyRevenue - serverCostEUR

	// ── Traffic alert (>90%) ──
	trafficAlert := false
	if usedPercent >= 90 {
		trafficAlert = true
		// Fire async admin TG alert (dedup via log — in production use a flag/cache)
		go notifyTrafficCritical(usedPercent, stats.TotalTraffic, serverLimitBytes)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"active_clients":     stats.ActiveClients,
		"total_upload":       stats.TotalUp,
		"total_download":     stats.TotalDown,
		"total_traffic":      stats.TotalTraffic,
		"server_limit_bytes": serverLimitBytes,
		"server_limit_gb":    serverLimitGB,
		"used_percent":       usedPercent,
		"traffic_alert":      trafficAlert,
		// Financial
		"monthly_revenue": monthlyRevenue,
		"server_cost":     serverCostEUR,
		"margin":          margin,
	})
}

// ══════════════════════════════════════════════════════════════
// GET /api/v1/admin/infra/vpn-active-clients — Active VPN users list
// Returns email, tariff, used GB, expiry for each active client.
// ══════════════════════════════════════════════════════════════

func AdminVPNActiveClientsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if GlobalDB == nil {
		json.NewEncoder(w).Encode(map[string]any{"error": "DB not ready"})
		return
	}

	// Fetch all completed VPN orders with meta
	rows, err := GlobalDB.Query(`
		SELECT o.provider_ref, o.product_name, o.price_usd, o.created_at,
		       COALESCE(o.meta, '{}'),
		       COALESCE(u.email, '')
		FROM store_orders o
		LEFT JOIN users u ON u.id = o.user_id
		WHERE o.status = 'completed'
		  AND o.provider_ref != ''
		  AND (o.activation_key LIKE 'vless://%' OR o.product_name ILIKE '%vpn%' OR o.product_name ILIKE '%безопасный%')
		ORDER BY o.created_at DESC
	`)
	if err != nil {
		log.Printf("[VPN-ACTIVE-CLIENTS] ❌ Query error: %v", err)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	defer rows.Close()

	type clientRow struct {
		Email        string  `json:"email"`
		ProductName  string  `json:"product_name"`
		PriceUSD     float64 `json:"price_usd"`
		CreatedAt    string  `json:"created_at"`
		ProviderRef  string  `json:"provider_ref"`
		TrafficBytes int64   `json:"traffic_bytes"`
		ExpireMs     int64   `json:"expire_ms"`
		DurationDays int     `json:"duration_days"`
		UsedBytes    int64   `json:"used_bytes"`
		UsedPercent  float64 `json:"used_percent"`
		Active       bool    `json:"active"`
	}

	// Get VlessProvider for live traffic queries
	var vp *vless.VlessProvider
	provider := shop.GetRegistry().Get("vless")
	if provider != nil {
		vp, _ = provider.(*vless.VlessProvider)
	}

	var clients []clientRow
	for rows.Next() {
		var ref, productName, email, metaStr string
		var priceUSD float64
		var createdAt string
		if err := rows.Scan(&ref, &productName, &priceUSD, &createdAt, &metaStr, &email); err != nil {
			continue
		}

		var meta struct {
			TrafficBytes int64 `json:"traffic_bytes"`
			ExpireMs     int64 `json:"expire_ms"`
			DurationDays int   `json:"duration_days"`
		}
		json.Unmarshal([]byte(metaStr), &meta)

		// Query live traffic
		var usedBytes int64
		if vp != nil {
			if stats, err := vp.GetClientTraffic(ref); err == nil {
				usedBytes = stats.Up + stats.Down
			}
		}

		usedPct := float64(0)
		if meta.TrafficBytes > 0 {
			usedPct = float64(usedBytes) / float64(meta.TrafficBytes) * 100
			if usedPct > 100 {
				usedPct = 100
			}
		}

		// Consider active if not expired and traffic not exhausted
		nowMs := time.Now().UnixMilli()
		active := (meta.ExpireMs == 0 || nowMs < meta.ExpireMs) && (meta.TrafficBytes == 0 || usedBytes < meta.TrafficBytes)

		clients = append(clients, clientRow{
			Email:        email,
			ProductName:  productName,
			PriceUSD:     priceUSD,
			CreatedAt:    createdAt,
			ProviderRef:  ref,
			TrafficBytes: meta.TrafficBytes,
			ExpireMs:     meta.ExpireMs,
			DurationDays: meta.DurationDays,
			UsedBytes:    usedBytes,
			UsedPercent:  usedPct,
			Active:       active,
		})
	}
	if clients == nil {
		clients = []clientRow{}
	}

	json.NewEncoder(w).Encode(map[string]any{"clients": clients})
}

// ══════════════════════════════════════════════════════════════
// Traffic critical alert — sends TG notification to all admins
// when server traffic exceeds 90%.
// ══════════════════════════════════════════════════════════════

func notifyTrafficCritical(usedPct float64, totalTraffic, limitBytes int64) {
	usedGB := float64(totalTraffic) / (1024 * 1024 * 1024)
	limitGB := float64(limitBytes) / (1024 * 1024 * 1024)
	msg := fmt.Sprintf("🚨 <b>КРИТИЧЕСКИЙ АЛЕРТ: Трафик сервера</b>\n\n"+
		"Использовано: <b>%.1f%%</b> (%.1f / %.0f ГБ)\n\n"+
		"⚠️ Серверный лимит трафика почти исчерпан.\n"+
		"Рекомендуется увеличить лимит или ограничить новые подключения.",
		usedPct, usedGB, limitGB)
	telegram.NotifyAdmins(msg, "", "")
}

// ══════════════════════════════════════════════════════════════
// NotifyAdminVPNPurchase — sends TG notification to admins
// after each VPN purchase with order details + monthly margin.
// Called from notifyStorePurchase when product is VPN.
// ══════════════════════════════════════════════════════════════

func NotifyAdminVPNPurchase(productName string, priceUSD string, userEmail string) {
	var monthlyRevenue float64
	var activeCount int

	if GlobalDB != nil {
		// Monthly revenue
		_ = GlobalDB.QueryRow(`
			SELECT COALESCE(SUM(price_usd), 0)
			FROM store_orders
			WHERE status = 'completed'
			  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%')
			  AND created_at >= date_trunc('month', NOW())
		`).Scan(&monthlyRevenue)

		// Active VPN user count (distinct users with completed VPN orders)
		_ = GlobalDB.QueryRow(`
			SELECT COUNT(DISTINCT user_id)
			FROM store_orders
			WHERE status = 'completed'
			  AND (activation_key LIKE 'vless://%' OR product_name ILIKE '%vpn%' OR product_name ILIKE '%безопасный%')
		`).Scan(&activeCount)
	}

	serverCost := 4.94
	if envCost := os.Getenv("VPN_SERVER_MONTHLY_COST"); envCost != "" {
		if v, err := strconv.ParseFloat(envCost, 64); err == nil && v > 0 {
			serverCost = v
		}
	}
	margin := monthlyRevenue - serverCost

	marginEmoji := "📈"
	if margin < 0 {
		marginEmoji = "📉"
	}

	msg := fmt.Sprintf("🔐 <b>Новая VPN-покупка</b>\n\n"+
		"👤 %s\n"+
		"📦 %s\n"+
		"💰 €%s\n\n"+
		"<b>Месячная аналитика:</b>\n"+
		"Выручка: €%.2f\n"+
		"Затраты: €%.2f\n"+
		"%s Маржа: €%.2f\n\n"+
		"👥 Всего VPN-пользователей: <b>%d</b>",
		userEmail, productName, priceUSD,
		monthlyRevenue, serverCost, marginEmoji, margin,
		activeCount)

	telegram.NotifyAdmins(msg, "Открыть админку", "https://xplr.pro/staff-only-zone")
}
