package wallet

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	useCase WalletUseCase
}

func NewHandler(uc WalletUseCase) *Handler {
	return &Handler{useCase: uc}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/wallet", func(r chi.Router) {
		r.Get("/balance", h.GetBalance)
		r.Post("/topup", h.TopUp)
		r.Post("/autotopup", h.ToggleAutoTopUp) // ← глобальный автотопап
	})
}

// GetBalance — GET /v1/wallet/balance.
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	balance, err := h.useCase.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{"balance": balance})
}

// TopUp — POST /v1/wallet/topup.
func (h *Handler) TopUp(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	type request struct {
		Amount float64 `json:"amount"`
	}

	var req request
	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.useCase.TopUpWallet(r.Context(), userID, domain.NewNumeric(req.Amount))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// ToggleAutoTopUp — POST /v1/wallet/autotopup.
func (h *Handler) ToggleAutoTopUp(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	type request struct {
		Enabled bool `json:"enabled"`
	}

	var req request
	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.useCase.ToggleAutoTopUp(r.Context(), userID, req.Enabled)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
