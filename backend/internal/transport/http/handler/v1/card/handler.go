package card

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
		r.Post("/", h.BuyCard)
		r.Get("/", h.ListCards)
		r.Get("/{cardID}", h.GetCard)
		r.Post("/{cardID}/topup", h.TopUpCard)
		r.Post("/{cardID}/close", h.CloseCard)
	})
}

// BuyCard — POST /v1/card
func (h *Handler) BuyCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	type request struct {
		Type string `json:"type"`
	}

	var req request
	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	cardType := domain.CardType(req.Type)
	card, err := h.useCase.BuyCard(r.Context(), userID, cardType)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, card)
}

// ListCards — GET /v1/card
func (h *Handler) ListCards(w http.ResponseWriter, r *http.Request) {
	// TODO: добавить ListByUserID
	handler.WriteJSON(w, http.StatusOK, []string{"TODO: ListCards"})
}

// GetCard — GET /v1/card/{cardID}
func (h *Handler) GetCard(w http.ResponseWriter, r *http.Request) {
	cardIDStr := chi.URLParam(r, "cardID")
	cardID, parseErr := uuid.Parse(cardIDStr)
	if parseErr != nil {
		http.Error(w, "invalid card_id", http.StatusBadRequest)
		return
	}

	card, err := h.useCase.GetByID(r.Context(), domain.UUID(cardID))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusNotFound)
		return
	}

	handler.WriteJSON(w, http.StatusOK, card)
}

// TopUpCard — POST /v1/card/{cardID}/topup
func (h *Handler) TopUpCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	cardIDStr := chi.URLParam(r, "cardID")

	cardID, _ := uuid.Parse(cardIDStr)

	type request struct {
		Amount float64 `json:"amount"`
	}

	var req request
	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.useCase.TopUpCard(r.Context(), userID, domain.UUID(cardID), domain.NewNumeric(req.Amount))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// CloseCard — POST /v1/card/{cardID}/close
func (h *Handler) CloseCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	cardIDStr := chi.URLParam(r, "cardID")

	cardID, _ := uuid.Parse(cardIDStr)

	err := h.useCase.CloseCard(r.Context(), userID, domain.UUID(cardID))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}
