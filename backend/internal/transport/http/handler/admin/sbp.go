package admin

import (
	"net/http"
	"strings"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterSBP(r chi.Router) {
	r.Get("/sbp-topup", h.GetSBPTopup)
	r.Patch("/sbp-topup", h.PatchSBPTopup)
}

func (h *Handler) GetSBPTopup(w http.ResponseWriter, r *http.Request) {
	enabled, err := h.sbpTopupEnabled(r)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]any{
		"enabled": enabled,
	})
}

// PatchSBPTopup — PATCH /admin/sbp-topup { enabled: bool }.
// Управляет system_settings.sbp_topup_enabled (setting_bool).
func (h *Handler) PatchSBPTopup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	row := &domain.SystemSetting{
		Key:         "sbp_topup_enabled",
		Value:       "",
		BoolValue:   &req.Enabled,
		Description: "Пополнение через СБП включено/отключено (boolean).",
	}

	err := h.systemRepo.Upsert(r.Context(), row)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]any{
		"status":  "success",
		"enabled": req.Enabled,
	})
}

func (h *Handler) sbpTopupEnabled(r *http.Request) (bool, error) {
	list, err := h.systemRepo.ListAll(r.Context())
	if err != nil {
		return false, wrapper.Wrap(err)
	}

	for _, row := range list {
		if row == nil || row.Key != "sbp_topup_enabled" {
			continue
		}
		if row.BoolValue != nil {
			return *row.BoolValue, nil
		}
		value := strings.TrimSpace(strings.ToLower(row.Value))
		if value == "0" || value == "false" || value == "off" {
			return false, nil
		}

		return true, nil
	}

	return true, nil
}
