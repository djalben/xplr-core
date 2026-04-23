package admin

import (
	"net/http"
	"strings"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterExchangeRates(r chi.Router) {
	r.Get("/exchange-rates", h.ListExchangeRates)
	r.Patch("/exchange-rates", h.PatchExchangeRate)
}

func (h *Handler) ListExchangeRates(w http.ResponseWriter, r *http.Request) {
	list, err := h.exchangeRateRepo.ListAll(r.Context())
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) PatchExchangeRate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrencyFrom  string  `json:"currency_from"`
		CurrencyTo    string  `json:"currency_to"`
		BaseRate      float64 `json:"base_rate"`
		MarkupPercent float64 `json:"markup_percent"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	cf := strings.ToUpper(strings.TrimSpace(req.CurrencyFrom))
	ct := strings.ToUpper(strings.TrimSpace(req.CurrencyTo))
	if cf == "" || ct == "" {
		http.Error(w, "currency_from and currency_to are required", http.StatusBadRequest)

		return
	}
	if req.BaseRate <= 0 {
		http.Error(w, "base_rate must be > 0", http.StatusBadRequest)

		return
	}

	base := domain.NewNumeric(req.BaseRate)
	markup := domain.NewNumeric(req.MarkupPercent)
	final := base
	if req.MarkupPercent != 0 {
		final = base.Mul(domain.NewNumeric(1.0 + req.MarkupPercent/100.0))
	}

	er := &domain.ExchangeRate{
		ID:            domain.NewUUID(),
		CurrencyFrom:  cf,
		CurrencyTo:    ct,
		BaseRate:      base,
		MarkupPercent: markup,
		FinalRate:     final,
	}

	err := h.exchangeRateRepo.Upsert(r.Context(), er)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
