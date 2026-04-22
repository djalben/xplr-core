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
// Управляет commission_config.sbp_topup_enabled: 1.0 (enabled) / 0.0 (disabled).
func (h *Handler) PatchSBPTopup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	cfg, err := h.commissionUseCase.GetByKey(r.Context(), "sbp_topup_enabled")
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	if req.Enabled {
		cfg.Value = domain.NewNumeric(1.0)
	} else {
		cfg.Value = domain.NewNumeric(0.0)
	}

	err = h.commissionUseCase.Update(r.Context(), cfg)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  "success",
		"enabled": req.Enabled,
	})
}
