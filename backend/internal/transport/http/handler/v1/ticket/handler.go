package ticket

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	useCase TicketUseCase
}

func NewHandler(uc TicketUseCase) *Handler {
	return &Handler{useCase: uc}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/ticket", func(r chi.Router) {
		r.Post("/create", h.Create)
		r.Put("/{id}/take", h.Take)
		r.Put("/{id}/close", h.Close)
	})
}

// Create — POST /v1/ticket/create.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	type request struct {
		Subject  string `json:"subject"`
		Message  string `json:"message"`
		TGChatID *int64 `json:"tgChatId,omitempty"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	ticket, err := h.useCase.Create(r.Context(), userID, req.Subject, req.Message, req.TGChatID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, ticket)
}

// Take — PUT /v1/ticket/{id}/take (для админа).
func (h *Handler) Take(w http.ResponseWriter, r *http.Request) {
	adminID := handler.GetUserIDFromContext(r) // потом проверим, что админ

	idStr := chi.URLParam(r, "id")
	id, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.useCase.Take(r.Context(), id, adminID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// Close — PUT /v1/ticket/{id}/close (для админа).
func (h *Handler) Close(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	type request struct {
		Reply string `json:"reply"`
	}

	var req request

	err = handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.useCase.Close(r.Context(), id, req.Reply)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
