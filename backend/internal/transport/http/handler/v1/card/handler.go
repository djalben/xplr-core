package card

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	useCase *card.UseCase
}

func NewHandler(uc *card.UseCase) *Handler {
	return &Handler{useCase: uc}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/card", func(r chi.Router) {
		r.Post("/buy", h.BuyCard)
		r.Post("/{id}/topup", h.TopUpCard)
		r.Put("/{id}/autotop", h.ToggleAutoTopUp)
	})
}

// BuyCard — POST /v1/card/buy
func (h *Handler) BuyCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	type request struct {
		CardType string `json:"card_type"` // subscriptions / travel / premium
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	card, err := h.useCase.BuyCard(r.Context(), userID, domain.CardType(req.CardType))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, card)
}

// TopUpCard — POST /v1/card/{id}/topup
func (h *Handler) TopUpCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	cardIDStr := chi.URLParam(r, "id")
	cardID, err := domain.ParseUUID(cardIDStr) // предположим, что есть ParseUUID в domain/types.go
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

// ToggleAutoTopUp — PUT /v1/card/{id}/autotop
func (h *Handler) ToggleAutoTopUp(w http.ResponseWriter, r *http.Request) {
	cardIDStr := chi.URLParam(r, "id")
	cardID, err := domain.ParseUUID(cardIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	type request struct {
		Enabled bool    `json:"enabled"`
		Below   float64 `json:"below"`
		Amount  float64 `json:"amount"`
	}

	var req request

	err = handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.useCase.ToggleAutoTopUp(r.Context(), cardID, req.Enabled, domain.NewNumeric(req.Below), domain.NewNumeric(req.Amount))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}