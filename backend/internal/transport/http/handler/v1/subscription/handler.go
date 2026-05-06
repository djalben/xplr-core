package subscription

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	uc SubscriptionUseCase
}

func NewHandler(uc SubscriptionUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/", h.List)
		r.Patch("/{id}", h.SetBlocked)
		r.Post("/block-by-card/{cardId}", h.SetBlockedByCard)
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	list, err := h.uc.List(r.Context(), userID)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)
		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]any{"subscriptions": list})
}

func (h *Handler) SetBlocked(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	idStr := chi.URLParam(r, "id")

	subID, err := domain.ParseUUID(idStr)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Неверный id")
		return
	}

	type req struct {
		IsBlocked bool `json:"is_blocked"`
	}

	var body req
	err = handler.ReadJSON(r, &body)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Неверный запрос")
		return
	}

	err = h.uc.SetBlocked(r.Context(), userID, subID, body.IsBlocked)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)
		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) SetBlockedByCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	idStr := chi.URLParam(r, "cardId")

	cardID, err := domain.ParseUUID(idStr)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Неверный id карты")
		return
	}

	type req struct {
		IsBlocked bool `json:"is_blocked"`
	}

	var body req
	err = handler.ReadJSON(r, &body)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Неверный запрос")
		return
	}

	err = h.uc.SetBlockedByCard(r.Context(), userID, cardID, body.IsBlocked)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)
		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]any{"ok": true})
}

