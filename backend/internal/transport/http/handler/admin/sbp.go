package admin

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterSBP(r chi.Router) {
	r.Patch("/sbp-topup", h.PatchSBPTopup)
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
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  "success",
		"enabled": req.Enabled,
	})
}
