package admin

import (
	"net/http"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterCommissions(r chi.Router) {
	r.Get("/commissions", h.ListCommissions)
	r.Patch("/commissions/{id}", h.PatchCommission)
}

func (h *Handler) ListCommissions(w http.ResponseWriter, r *http.Request) {
	list, err := h.commissionUseCase.ListAll(r.Context())
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, list)
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
		_ = handler.WriteInternalServerError(r.Context(), w, err)

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

	err = h.commissionUseCase.Update(r.Context(), cfg)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Неверные данные")

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]string{"status": "success"})
}
