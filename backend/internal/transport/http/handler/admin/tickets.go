package admin

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterTickets(r chi.Router) {
	r.Get("/tickets", h.ListTickets)
}

func (h *Handler) ListTickets(w http.ResponseWriter, r *http.Request) {
	list, err := h.ticketUseCase.ListAll(r.Context(), 200)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, list)
}
