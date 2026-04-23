package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterInfra(r chi.Router) {
	r.Route("/infra", func(r chi.Router) {
		r.Get("/balance", h.InfraBalance)
		r.Get("/server-info", h.InfraServerInfo)
		r.Get("/active-keys", h.InfraActiveKeys)
		r.Get("/vpn-server-status", h.InfraVPNServerStatus)
		r.Get("/vpn-active-clients", h.InfraVPNActiveClients)
	})

	r.Route("/vpn", func(r chi.Router) {
		r.Delete("/client/{providerRef}", h.DeleteVPNClient)
		r.Patch("/client/{providerRef}", h.PatchVPNClient)
	})
}

func (h *Handler) settingsMap(ctx context.Context) (map[string]string, error) {
	list, err := h.systemRepo.ListAll(ctx)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	out := make(map[string]string, len(list))
	for _, row := range list {
		if row == nil {
			continue
		}
		out[row.Key] = row.Value
	}

	return out, nil
}

func (h *Handler) InfraBalance(w http.ResponseWriter, r *http.Request) {
	// In main this comes from Aeza; we mirror shape and allow seeding via system_settings.
	m, err := h.settingsMap(r.Context())
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	bal, _ := strconv.ParseFloat(strings.TrimSpace(m["aeza_balance"]), 64)
	cur := strings.TrimSpace(m["aeza_currency"])
	if cur == "" {
		cur = "EUR"
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"balance":    bal,
		"currency":   cur,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) InfraServerInfo(w http.ResponseWriter, r *http.Request) {
	m, err := h.settingsMap(r.Context())
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	// Default placeholders; can be overridden by system_settings.
	id, _ := strconv.Atoi(strings.TrimSpace(m["vpn_server_id"]))
	if id == 0 {
		id = 0
	}

	cost, _ := strconv.ParseFloat(strings.TrimSpace(m["vpn_server_cost_eur"]), 64)
	cpu, _ := strconv.Atoi(strings.TrimSpace(m["vpn_server_cpu"]))
	ramMb, _ := strconv.Atoi(strings.TrimSpace(m["vpn_server_ram_mb"]))
	diskGb, _ := strconv.Atoi(strings.TrimSpace(m["vpn_server_disk_gb"]))

	out := map[string]any{
		"id":         id,
		"name":       strings.TrimSpace(m["vpn_server_name"]),
		"status":     strings.TrimSpace(m["vpn_server_status"]),
		"ip":         strings.TrimSpace(m["vpn_server_ip"]),
		"cost_eur":   cost,
		"expires_at": strings.TrimSpace(m["vpn_server_expires_at"]),
		"cpu":        cpu,
		"ram_mb":     ramMb,
		"disk_gb":    diskGb,
		"disk_type":  strings.TrimSpace(m["vpn_server_disk_type"]),
		"os":         strings.TrimSpace(m["vpn_server_os"]),
		"location":   strings.TrimSpace(m["vpn_server_location"]),
		"api_status": "ok",
	}

	if out["name"] == "" && out["ip"] == "" && cost == 0 {
		out["api_status"] = "error"
	}

	handler.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) InfraActiveKeys(w http.ResponseWriter, r *http.Request) {
	clients, err := h.listVPNClients(r.Context(), 2000)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	active := 0
	for _, c := range clients {
		if c.Active {
			active++
		}
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{"active_keys": active})
}

func (h *Handler) InfraVPNActiveClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.listVPNClients(r.Context(), 2000)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{"clients": clients})
}

type vpnClientRow struct {
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

func (h *Handler) listVPNClients(ctx context.Context, limit int) ([]*vpnClientRow, error) {
	rows, err := h.storeRepo.AdminListVPNOrders(ctx, limit)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	out := make([]*vpnClientRow, 0, len(rows))

	for _, r := range rows {
		if r == nil {
			continue
		}

		meta := struct {
			TrafficBytes int64 `json:"traffic_bytes"`
			ExpireMs     int64 `json:"expire_ms"`
			DurationDays int   `json:"duration_days"`
		}{}

		_ = json.Unmarshal([]byte(r.Meta), &meta)

		usedBytes := int64(0)
		active := false
		if h.vpnAdminProvider != nil && r.ProviderRef != "" {
			tr, err := h.vpnAdminProvider.GetClientTraffic(ctx, r.ProviderRef)
			if err == nil && tr != nil {
				usedBytes = tr.Upload + tr.Download
				active = tr.Enabled
			}
		}

		usedPct := 0.0
		if meta.TrafficBytes > 0 {
			usedPct = (float64(usedBytes) / float64(meta.TrafficBytes)) * 100.0
		}

		out = append(out, &vpnClientRow{
			Email:         r.UserEmail,
			FullName:      "",
			ProductName:   r.ProductName,
			PriceUSD:      r.PriceUSD,
			CreatedAt:     r.CreatedAt.UTC().Format(time.RFC3339),
			ProviderRef:   r.ProviderRef,
			ActivationKey: r.ActivationKey,
			TrafficBytes:  meta.TrafficBytes,
			ExpireMs:      meta.ExpireMs,
			DurationDays:  meta.DurationDays,
			UsedBytes:     usedBytes,
			UsedPercent:   usedPct,
			Active:        active,
		})
	}

	return out, nil
}

func (h *Handler) InfraVPNServerStatus(w http.ResponseWriter, r *http.Request) {
	clients, err := h.listVPNClients(r.Context(), 5000)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	m, err := h.settingsMap(r.Context())
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	limitGB, _ := strconv.ParseFloat(strings.TrimSpace(m["vpn_server_limit_gb"]), 64)
	if limitGB <= 0 {
		limitGB = 30
	}
	limitBytes := int64(limitGB * 1024 * 1024 * 1024)

	totalTraffic := int64(0)
	totalUp := int64(0)
	totalDown := int64(0)
	activeCount := 0
	uniq := map[string]struct{}{}
	for _, c := range clients {
		if c == nil {
			continue
		}
		totalTraffic += c.UsedBytes
		// We only have used_bytes; split as download for display parity.
		totalDown += c.UsedBytes
		if c.Active {
			activeCount++
		}
		if c.ProviderRef != "" {
			uniq[c.ProviderRef] = struct{}{}
		}
	}

	usedPct := 0.0
	if limitBytes > 0 {
		usedPct = (float64(totalTraffic) / float64(limitBytes)) * 100
	}
	trafficAlert := usedPct >= 80

	// Monthly revenue and clients based on orders timing (created_at in rows).
	now := time.Now().UTC()
	startMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startPrev := startMonth.AddDate(0, -1, 0)

	monthlyRevenue := 0.0
	currentMonthClients := 0
	prevMonthClients := 0
	for _, c := range clients {
		if c == nil {
			continue
		}
		t, err := time.Parse(time.RFC3339, c.CreatedAt)
		if err != nil {
			continue
		}
		if !t.Before(startMonth) {
			monthlyRevenue += c.PriceUSD
			currentMonthClients++

			continue
		}
		if !t.Before(startPrev) {
			prevMonthClients++
		}
	}

	serverCost, _ := strconv.ParseFloat(strings.TrimSpace(m["vpn_server_cost_eur"]), 64)
	margin := monthlyRevenue - serverCost

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"active_clients":        activeCount,
		"total_upload":          totalUp,
		"total_download":        totalDown,
		"total_traffic":         totalTraffic,
		"server_limit_bytes":    limitBytes,
		"server_limit_gb":       limitGB,
		"used_percent":          usedPct,
		"traffic_alert":         trafficAlert,
		"monthly_revenue":       monthlyRevenue,
		"server_cost":           serverCost,
		"margin":                margin,
		"unique_vpn_clients":    len(uniq),
		"current_month_clients": currentMonthClients,
		"prev_month_clients":    prevMonthClients,
	})
}

func (h *Handler) DeleteVPNClient(w http.ResponseWriter, r *http.Request) {
	providerRef := chi.URLParam(r, "providerRef")
	providerRef = strings.TrimSpace(providerRef)
	if providerRef == "" {
		http.Error(w, "providerRef is required", http.StatusBadRequest)

		return
	}
	if h.vpnAdminProvider == nil {
		http.Error(w, "vpn provider is not configured", http.StatusBadRequest)

		return
	}

	err := h.vpnAdminProvider.DeleteClientByEmail(r.Context(), providerRef)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	// Also mark related orders as deleted in our DB.
	_ = h.storeRepo.SoftDeleteOrdersByProviderRef(r.Context(), providerRef)

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) PatchVPNClient(w http.ResponseWriter, r *http.Request) {
	providerRef := chi.URLParam(r, "providerRef")
	providerRef = strings.TrimSpace(providerRef)
	if providerRef == "" {
		http.Error(w, "providerRef is required", http.StatusBadRequest)

		return
	}
	if h.vpnAdminProvider == nil {
		http.Error(w, "vpn provider is not configured", http.StatusBadRequest)

		return
	}

	var req struct {
		TotalBytes *int64 `json:"total_bytes"`
		ExpiryMs   *int64 `json:"expiry_ms"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	err := h.vpnAdminProvider.UpdateClientByEmail(r.Context(), providerRef, req.TotalBytes, req.ExpiryMs)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	// Keep meta in sync best-effort.
	metaStr, err := h.storeRepo.GetLatestCompletedOrderMetaByProviderRef(r.Context(), providerRef, nil)
	if err == nil && metaStr != "" {
		var meta map[string]any
		_ = json.Unmarshal([]byte(metaStr), &meta)
		if req.TotalBytes != nil {
			meta["traffic_bytes"] = *req.TotalBytes
		}
		if req.ExpiryMs != nil {
			meta["expire_ms"] = *req.ExpiryMs
		}
		b, err := json.Marshal(meta)
		if err == nil && len(b) > 0 {
			_ = h.storeRepo.UpdateOrderMetaByProviderRef(r.Context(), providerRef, string(b))
		}
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
