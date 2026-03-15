package transaction

import (
	"net/http"
	"time"

	"github.com/djalben/xplr-core/backend/internal/application/transaction"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	useCase *transaction.UseCase
}

func NewHandler(uc *transaction.UseCase) *Handler {
	return &Handler{useCase: uc}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/transaction", func(r chi.Router) {
		r.Get("/wallet", h.GetWalletTransactions)
		r.Get("/card/{id}", h.GetCardTransactions)
	})
}

// GetWalletTransactions — GET /v1/transaction/wallet?from=&to=.
func (h *Handler) GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		from = time.Now().AddDate(0, -1, 0) // дефолт месяц назад
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		to = time.Now()
	}

	txs, err := h.useCase.GetWalletTransactions(r.Context(), userID, from, to)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, txs)
}

// GetCardTransactions — GET /v1/transaction/card/{id}?from=&to=.
func (h *Handler) GetCardTransactions(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		from = time.Now().AddDate(0, -1, 0)
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		to = time.Now()
	}

	txs, err := h.useCase.GetCardTransactions(r.Context(), id, from, to)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, txs)
}
