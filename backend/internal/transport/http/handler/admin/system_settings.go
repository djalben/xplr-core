package admin

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterSystemSettings(r chi.Router) {
	r.Get("/system-settings", h.ListSystemSettings)
	r.Patch("/system-settings/{key}", h.PatchSystemSetting)
}

func (h *Handler) ListSystemSettings(w http.ResponseWriter, r *http.Request) {
	list, err := h.systemRepo.ListAll(r.Context())
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, list)
}

func (h *Handler) PatchSystemSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)

		return
	}

	var req struct {
		Value       string  `json:"value"`
		Description *string `json:"description"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	s := &domain.SystemSetting{
		Key:   key,
		Value: req.Value,
	}
	if req.Description != nil {
		s.Description = *req.Description
	}

	err := h.systemRepo.Upsert(r.Context(), s)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]string{"status": "success"})
}
