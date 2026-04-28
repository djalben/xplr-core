package admin

import (
	"net/http"
	"strconv"

	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterLogs(r chi.Router) {
	r.Get("/logs", h.ListAdminLogs)
}

func (h *Handler) ListAdminLogs(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if s := r.URL.Query().Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			limit = v
		}
	}

	list, err := h.adminLogsRepo.List(r.Context(), limit)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, list)
}
