package admin

import (
	"net/http"
	"strconv"

	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterLogs(r chi.Router) {
	r.Get("/logs", h.ListAdminLogs)
}

func (h *Handler) ListAdminLogs(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if s := r.URL.Query().Get("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			limit = v
		}
	}

	list, err := h.adminLogsRepo.List(r.Context(), limit)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, list)
}
