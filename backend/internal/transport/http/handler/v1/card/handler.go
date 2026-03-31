package card

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	useCase CardUseCase
}

func NewHandler(uc CardUseCase) *Handler {
	return &Handler{useCase: uc}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/card", func(r chi.Router) {
		r.Post("/buy", h.BuyCard)
		r.Post("/{id}/topup", h.TopUpCard)
		r.Post("/{id}/spend", h.SpendFromCard)
	})
}

// BuyCard — POST /v1/card/buy.
func (h *Handler) BuyCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	type request struct {
		CardType string `json:"cardType"`
		Nickname string `json:"nickname"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	card, err := h.useCase.BuyCard(r.Context(), userID, domain.CardType(req.CardType), req.Nickname)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, card)
}

// TopUpCard — POST /v1/card/{id}/topup.
func (h *Handler) TopUpCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	cardIDStr := chi.URLParam(r, "id")

	cardID, err := domain.ParseUUID(cardIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	type request struct {
		Amount float64 `json:"amount"`
	}

	var req request

	err = handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.useCase.TopUpCard(r.Context(), userID, cardID, domain.NewNumeric(req.Amount))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// SpendFromCard — POST /v1/card/{id}/spend.
func (h *Handler) SpendFromCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	cardIDStr := chi.URLParam(r, "id")

	cardID, err := domain.ParseUUID(cardIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	type request struct {
		Amount float64 `json:"amount"`
	}

	var req request

	err = handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.useCase.SpendFromCard(r.Context(), userID, cardID, domain.NewNumeric(req.Amount))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
