package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/providers/vless"
	"github.com/djalben/xplr-core/backend/shop"
	"github.com/djalben/xplr-core/backend/telegram"
	"github.com/gorilla/mux"
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

	// Server traffic limit from env (in GB, default 30 GB — real Aeza plan)
	serverLimitGB := 30
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
	serverCostEUR := 4.94
	if envCost := os.Getenv("VPN_SERVER_MONTHLY_COST"); envCost != "" {
		if v, err := strconv.ParseFloat(envCost, 64); err == nil && v > 0 {
			serverCostEUR = v
		}
	}

	// Revenue: sum of all completed VPN orders this month
	var monthlyRevenue float64
	var uniqueVPNClients int
	var currentMonthClients int
	var prevMonthClients int
	if GlobalDB != nil {
		_ = GlobalDB.QueryRow(`
			SELECT COALESCE(SUM(price_usd), 0)
			FROM store_orders
			WHERE status = 'completed'
			  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%')
			  AND created_at >= date_trunc('month', NOW())
		`).Scan(&monthlyRevenue)

		// Unique VPN clients (users with at least one active=completed VPN order)
		_ = GlobalDB.QueryRow(`
			SELECT COUNT(DISTINCT user_id)
			FROM store_orders
			WHERE status = 'completed'
			  AND (activation_key LIKE 'vless://%' OR product_name ILIKE '%vpn%' OR product_name ILIKE '%безопасный%')
		`).Scan(&uniqueVPNClients)

		// Current month new VPN clients
		_ = GlobalDB.QueryRow(`
			SELECT COUNT(DISTINCT user_id)
			FROM store_orders
			WHERE status = 'completed'
			  AND (activation_key LIKE 'vless://%' OR product_name ILIKE '%vpn%' OR product_name ILIKE '%безопасный%')
			  AND created_at >= date_trunc('month', NOW())
		`).Scan(&currentMonthClients)

		// Previous month VPN clients
		_ = GlobalDB.QueryRow(`
			SELECT COUNT(DISTINCT user_id)
			FROM store_orders
			WHERE status = 'completed'
			  AND (activation_key LIKE 'vless://%' OR product_name ILIKE '%vpn%' OR product_name ILIKE '%безопасный%')
			  AND created_at >= date_trunc('month', NOW()) - INTERVAL '1 month'
			  AND created_at < date_trunc('month', NOW())
		`).Scan(&prevMonthClients)
	}

	margin := monthlyRevenue - serverCostEUR

	// ── Traffic alert: critical when remaining < 10 GB ──
	remainingBytes := serverLimitBytes - stats.TotalTraffic
	if remainingBytes < 0 {
		remainingBytes = 0
	}
	remainingGB := float64(remainingBytes) / (1024 * 1024 * 1024)
	trafficAlert := false
	if remainingGB < 10 {
		trafficAlert = true
		go notifyTrafficCritical(usedPercent, stats.TotalTraffic, serverLimitBytes)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"active_clients":        stats.ActiveClients,
		"total_upload":          stats.TotalUp,
		"total_download":        stats.TotalDown,
		"total_traffic":         stats.TotalTraffic,
		"server_limit_bytes":    serverLimitBytes,
		"server_limit_gb":       serverLimitGB,
		"used_percent":          usedPercent,
		"traffic_alert":         trafficAlert,
		"unique_vpn_clients":    uniqueVPNClients,
		"current_month_clients": currentMonthClients,
		"prev_month_clients":    prevMonthClients,
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

	// Fetch all completed VPN orders with meta, user name, and activation key
	rows, err := GlobalDB.Query(`
		SELECT o.provider_ref, o.product_name, o.price_usd, o.created_at,
		       COALESCE(o.meta, '{}'),
		       COALESCE(u.email, ''),
		       COALESCE(u.display_name, ''),
		       COALESCE(o.activation_key, '')
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
		Email         string  `json:"email"`
		FullName      string  `json:"full_name"`
		ProductName   string  `json:"product_name"`
		PriceUSD      float64 `json:"price_usd"`
		CreatedAt     string  `json:"created_at"`
		ProviderRef   string  `json:"provider_ref"`
		ActivationKey string  `json:"activation_key"`
		TrafficBytes  int64   `json:"traffic_bytes"`
		ExpireMs      int64   `json:"expire_ms"`
		DurationDays  int     `json:"duration_days"`
		UsedBytes     int64   `json:"used_bytes"`
		UsedPercent   float64 `json:"used_percent"`
		Active        bool    `json:"active"`
	}

	// Get VlessProvider for live traffic queries
	var vp *vless.VlessProvider
	provider := shop.GetRegistry().Get("vless")
	if provider != nil {
		vp, _ = provider.(*vless.VlessProvider)
	}

	var clients []clientRow
	for rows.Next() {
		var ref, productName, email, metaStr, fullName, activationKey string
		var priceUSD float64
		var createdAt string
		if err := rows.Scan(&ref, &productName, &priceUSD, &createdAt, &metaStr, &email, &fullName, &activationKey); err != nil {
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
			Email:         email,
			FullName:      fullName,
			ProductName:   productName,
			PriceUSD:      priceUSD,
			CreatedAt:     createdAt,
			ProviderRef:   ref,
			ActivationKey: activationKey,
			TrafficBytes:  meta.TrafficBytes,
			ExpireMs:      meta.ExpireMs,
			DurationDays:  meta.DurationDays,
			UsedBytes:     usedBytes,
			UsedPercent:   usedPct,
			Active:        active,
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
	remainingGB := limitGB - usedGB
	if remainingGB < 0 {
		remainingGB = 0
	}
	msg := fmt.Sprintf("🚨 <b>Внимание! Критический остаток трафика на сервере</b>\n\n"+
		"Остаток: <b>%.1f ГБ</b> из %.0f ГБ\n"+
		"Использовано: <b>%.1f%%</b> (%.1f ГБ)\n\n"+
		"⚠️ Рекомендуется увеличить лимит или ограничить новые подключения.",
		remainingGB, limitGB, usedPct, usedGB)
	telegram.NotifyAdmins(msg, "Открыть админку", "https://xplr.pro/staff-only-zone")
}

// ══════════════════════════════════════════════════════════════
// NotifyAdminVPNPurchase — sends TG notification to admins
// after each VPN purchase with order details + monthly margin.
// Called from notifyStorePurchase when product is VPN.
// ══════════════════════════════════════════════════════════════

func NotifyAdminVPNPurchase(productName string, priceUSD string, userEmail string) {
	log.Printf("[ADMIN-VPN-NOTIFY] 🔔 START: product=%q price=%s user=%s", productName, priceUSD, userEmail)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ADMIN-VPN-NOTIFY] ❌ PANIC: %v", r)
		}
	}()

	var monthlyRevenue float64
	var activeCount int

	if GlobalDB != nil {
		// Monthly revenue
		err := GlobalDB.QueryRow(`
			SELECT COALESCE(SUM(price_usd), 0)
			FROM store_orders
			WHERE status = 'completed'
			  AND (product_name ILIKE '%vpn%' OR product_name ILIKE '%vless%' OR product_name ILIKE '%безопасный%')
			  AND created_at >= date_trunc('month', NOW())
		`).Scan(&monthlyRevenue)
		if err != nil {
			log.Printf("[ADMIN-VPN-NOTIFY] ⚠️ Revenue query error: %v", err)
		}

		// Active VPN user count (distinct users with completed VPN orders, excluding deleted)
		err = GlobalDB.QueryRow(`
			SELECT COUNT(DISTINCT user_id)
			FROM store_orders
			WHERE status = 'completed'
			  AND (activation_key LIKE 'vless://%' OR product_name ILIKE '%vpn%' OR product_name ILIKE '%безопасный%')
		`).Scan(&activeCount)
		if err != nil {
			log.Printf("[ADMIN-VPN-NOTIFY] ⚠️ Active count query error: %v", err)
		}
		log.Printf("[ADMIN-VPN-NOTIFY] 📊 Revenue=€%.2f, ActiveCount=%d", monthlyRevenue, activeCount)
	} else {
		log.Printf("[ADMIN-VPN-NOTIFY] ⚠️ GlobalDB is nil — cannot query analytics")
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
		"👥 Всего активных пользователей: <b>%d</b>",
		userEmail, productName, priceUSD,
		monthlyRevenue, serverCost, marginEmoji, margin,
		activeCount)

	log.Printf("[ADMIN-VPN-NOTIFY] 📤 Calling telegram.NotifyAdmins (msg length=%d)...", len(msg))
	telegram.NotifyAdmins(msg, "Открыть админку", "https://xplr.pro/staff-only-zone")
	log.Printf("[ADMIN-VPN-NOTIFY] ✅ telegram.NotifyAdmins call completed for %s", userEmail)
}

// ══════════════════════════════════════════════════════════════
// DELETE /api/v1/admin/vpn/client/{email}
// Deletes client from 3X-UI panel first, then removes DB record.
// ══════════════════════════════════════════════════════════════

func AdminDeleteVPNClientHandler(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	if email == "" {
		http.Error(w, `{"error":"missing email"}`, http.StatusBadRequest)
		return
	}

	log.Printf("[VPN-DELETE] 🗑 Deleting client %s...", email)

	// 1. Look up the client UUID from the activation_key (vless://UUID@...)
	var activationKey string
	err := GlobalDB.QueryRow(`
		SELECT COALESCE(activation_key, '')
		FROM store_orders
		WHERE provider_ref = $1 AND status = 'completed'
		ORDER BY created_at DESC LIMIT 1
	`, email).Scan(&activationKey)
	if err != nil || activationKey == "" {
		log.Printf("[VPN-DELETE] ❌ Order not found for provider_ref=%q, err=%v, key=%q", email, err, activationKey)
		http.Error(w, `{"error":"order not found"}`, http.StatusNotFound)
		return
	}

	preview := activationKey
	if len(preview) > 120 {
		preview = preview[:120]
	}
	log.Printf("[VPN-DELETE] 🔍 Raw activation_key for %s: %q (len=%d)", email, preview, len(activationKey))

	// If activation_key looks like base64, try decoding it first
	keyToExtract := activationKey
	if !strings.Contains(activationKey, "://") && len(activationKey) > 40 {
		if decoded, err := base64.StdEncoding.DecodeString(activationKey); err == nil {
			decodedStr := string(decoded)
			if strings.Contains(strings.ToLower(decodedStr), "vless://") {
				log.Printf("[VPN-DELETE] 🔄 Decoded base64 activation_key: %q", decodedStr[:min(120, len(decodedStr))])
				keyToExtract = decodedStr
			}
		} else if decoded, err := base64.RawStdEncoding.DecodeString(activationKey); err == nil {
			decodedStr := string(decoded)
			if strings.Contains(strings.ToLower(decodedStr), "vless://") {
				log.Printf("[VPN-DELETE] 🔄 Decoded raw-base64 activation_key: %q", decodedStr[:min(120, len(decodedStr))])
				keyToExtract = decodedStr
			}
		}
	}

	// Extract UUID from vless://UUID@...
	clientUUID := extractUUIDFromVlessLink(keyToExtract)
	if clientUUID == "" {
		log.Printf("[VPN-DELETE] ❌ Cannot extract UUID. Raw key: %q", preview)
		http.Error(w, fmt.Sprintf(`{"error":"cannot extract UUID from activation key","activation_key_preview":"%s"}`, preview), http.StatusInternalServerError)
		return
	}

	// 2. Delete from 3X-UI panel FIRST
	provider := shop.GetRegistry().Get("vless")
	if provider == nil {
		http.Error(w, `{"error":"vless provider not registered"}`, http.StatusInternalServerError)
		return
	}
	vp, ok := provider.(*vless.VlessProvider)
	if !ok {
		http.Error(w, `{"error":"provider type assertion failed"}`, http.StatusInternalServerError)
		return
	}

	if err := vp.DeleteClient(clientUUID); err != nil {
		log.Printf("[VPN-DELETE] ❌ 3X-UI deleteClient failed for %s: %v", email, err)
		http.Error(w, fmt.Sprintf(`{"error":"3X-UI delete failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	log.Printf("[VPN-DELETE] ✅ Client %s removed from 3X-UI panel", email)

	// 3. Mark order as deleted in DB (soft-delete: set status = 'deleted')
	_, err = GlobalDB.Exec(`
		UPDATE store_orders SET status = 'deleted'
		WHERE provider_ref = $1 AND status = 'completed'
	`, email)
	if err != nil {
		log.Printf("[VPN-DELETE] ⚠️ DB update failed for %s: %v (client already removed from panel)", email, err)
	}

	log.Printf("[VPN-DELETE] ✅ Client %s fully deleted", email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"message": "Client deleted from panel and database",
	})
}

// ══════════════════════════════════════════════════════════════
// PATCH /api/v1/admin/vpn/client/{email}
// Updates client total_bytes and/or expiry_time on 3X-UI + DB.
// ══════════════════════════════════════════════════════════════

func AdminEditVPNClientHandler(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	if email == "" {
		http.Error(w, `{"error":"missing email"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		TotalBytes *int64 `json:"total_bytes"`
		ExpiryMs   *int64 `json:"expiry_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	if req.TotalBytes == nil && req.ExpiryMs == nil {
		http.Error(w, `{"error":"nothing to update"}`, http.StatusBadRequest)
		return
	}

	log.Printf("[VPN-EDIT] ✏️ Editing client %s: totalBytes=%v expiryMs=%v", email, req.TotalBytes, req.ExpiryMs)

	// 1. Fetch current order data
	var activationKey, metaStr string
	err := GlobalDB.QueryRow(`
		SELECT COALESCE(activation_key, ''), COALESCE(meta, '{}')
		FROM store_orders
		WHERE provider_ref = $1 AND status = 'completed'
		ORDER BY created_at DESC LIMIT 1
	`, email).Scan(&activationKey, &metaStr)
	if err != nil || activationKey == "" {
		http.Error(w, `{"error":"order not found"}`, http.StatusNotFound)
		return
	}

	// Try base64 decode if not a direct vless:// link
	keyToExtract := activationKey
	if !strings.Contains(activationKey, "://") && len(activationKey) > 40 {
		if decoded, decErr := base64.StdEncoding.DecodeString(activationKey); decErr == nil {
			if strings.Contains(strings.ToLower(string(decoded)), "vless://") {
				keyToExtract = string(decoded)
			}
		} else if decoded, decErr := base64.RawStdEncoding.DecodeString(activationKey); decErr == nil {
			if strings.Contains(strings.ToLower(string(decoded)), "vless://") {
				keyToExtract = string(decoded)
			}
		}
	}

	clientUUID := extractUUIDFromVlessLink(keyToExtract)
	if clientUUID == "" {
		log.Printf("[VPN-EDIT] ❌ Cannot extract UUID from key (len=%d): %q", len(activationKey), activationKey[:min(120, len(activationKey))])
		http.Error(w, `{"error":"cannot extract UUID from activation key"}`, http.StatusInternalServerError)
		return
	}

	// Parse existing meta
	var meta struct {
		TrafficBytes int64 `json:"traffic_bytes"`
		ExpireMs     int64 `json:"expire_ms"`
		DurationDays int   `json:"duration_days"`
	}
	json.Unmarshal([]byte(metaStr), &meta)

	// Apply updates
	newTotal := meta.TrafficBytes
	newExpiry := meta.ExpireMs
	if req.TotalBytes != nil {
		newTotal = *req.TotalBytes
	}
	if req.ExpiryMs != nil {
		newExpiry = *req.ExpiryMs
	}

	// 2. Update on 3X-UI panel
	provider := shop.GetRegistry().Get("vless")
	if provider == nil {
		http.Error(w, `{"error":"vless provider not registered"}`, http.StatusInternalServerError)
		return
	}
	vp, ok := provider.(*vless.VlessProvider)
	if !ok {
		http.Error(w, `{"error":"provider type assertion failed"}`, http.StatusInternalServerError)
		return
	}

	if err := vp.UpdateClient(clientUUID, email, newTotal, newExpiry); err != nil {
		log.Printf("[VPN-EDIT] ❌ 3X-UI updateClient failed for %s: %v", email, err)
		http.Error(w, fmt.Sprintf(`{"error":"3X-UI update failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}

	// 3. Update meta in DB
	meta.TrafficBytes = newTotal
	meta.ExpireMs = newExpiry
	newMetaJSON, _ := json.Marshal(meta)
	_, err = GlobalDB.Exec(`
		UPDATE store_orders SET meta = $1
		WHERE provider_ref = $2 AND status = 'completed'
	`, string(newMetaJSON), email)
	if err != nil {
		log.Printf("[VPN-EDIT] ⚠️ DB meta update failed for %s: %v", email, err)
	}

	log.Printf("[VPN-EDIT] ✅ Client %s updated: totalBytes=%d expiryMs=%d", email, newTotal, newExpiry)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":          true,
		"total_bytes": newTotal,
		"expiry_ms":   newExpiry,
	})
}

// extractUUIDFromVlessLink extracts the UUID from a vless://UUID@IP:PORT?... link.
// Handles extra slashes, URL-encoded chars, and complex query parameters.
func extractUUIDFromVlessLink(link string) string {
	link = strings.TrimSpace(link)
	preview := link
	if len(preview) > 120 {
		preview = preview[:120]
	}

	if link == "" {
		log.Printf("[UUID-PARSE] ❌ Empty activation key")
		return ""
	}

	// Case-insensitive prefix check
	lower := strings.ToLower(link)
	if !strings.HasPrefix(lower, "vless://") {
		log.Printf("[UUID-PARSE] ❌ Not a vless:// link: %q", preview)
		return ""
	}

	// Strip scheme and any extra leading slashes
	rest := link[8:] // skip "vless://"
	rest = strings.TrimLeft(rest, "/")

	// URL-decode the entire user-info part first
	if decoded, err := url.PathUnescape(rest); err == nil {
		rest = decoded
	}

	// The UUID is the part before the first '@'
	atIdx := strings.Index(rest, "@")
	if atIdx <= 0 {
		log.Printf("[UUID-PARSE] ❌ No '@' found after scheme in link: %q", preview)
		return ""
	}
	uuid := strings.TrimSpace(rest[:atIdx])

	// Strip any trailing slashes or whitespace from UUID
	uuid = strings.Trim(uuid, "/ \t\r\n")

	// Validate UUID format: 8-4-4-4-12 hex chars with dashes
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(uuid) {
		// Try to recover: strip all non-hex characters and reformat
		hexOnly := regexp.MustCompile(`[^0-9a-fA-F]`).ReplaceAllString(uuid, "")
		if len(hexOnly) == 32 {
			uuid = fmt.Sprintf("%s-%s-%s-%s-%s", hexOnly[:8], hexOnly[8:12], hexOnly[12:16], hexOnly[16:20], hexOnly[20:])
			log.Printf("[UUID-PARSE] ⚠️ Recovered UUID from raw hex: %s", uuid)
		} else {
			log.Printf("[UUID-PARSE] ❌ Invalid UUID format (len=%d, hex=%d): %q from link: %q", len(uuid), len(hexOnly), uuid, preview)
			return ""
		}
	}

	log.Printf("[UUID-PARSE] ✅ Extracted UUID: %s", uuid)
	return uuid
}
