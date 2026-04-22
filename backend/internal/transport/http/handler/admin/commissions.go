package admin

import (
	"net/http"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterCommissions(r chi.Router) {
	r.Get("/commissions", h.ListCommissions)
	r.Patch("/commissions/{id}", h.PatchCommission)
}

func (h *Handler) ListCommissions(w http.ResponseWriter, r *http.Request) {
	list, err := h.commissionUseCase.ListAll(r.Context())
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) PatchCommission(w http.ResponseWriter, r *http.Request) {
	id, ok := adminChiUUID(w, r)
	if !ok {
		return
	}

	var req struct {
		Value float64 `json:"value"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	list, err := h.commissionUseCase.ListAll(r.Context())
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	var cfg *domain.CommissionConfig
	for _, c := range list {
		if c.ID == id {
			cfg = c
			break
		}
	}
	if cfg == nil {
		http.Error(w, "not found", http.StatusNotFound)

		return
	}

	cfg.Value = domain.NewNumeric(req.Value)
	cfg.UpdatedAt = time.Now().UTC()

	if err := h.commissionUseCase.Update(r.Context(), cfg); err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
