package store

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/djalben/xplr-core/backend/internal/application/store"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	uc        *store.UseCase
	storeRepo ports.StoreRepository
}

func NewHandler(uc *store.UseCase, storeRepo ports.StoreRepository) *Handler {
	return &Handler{uc: uc, storeRepo: storeRepo}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/user/store", func(r chi.Router) {
		r.Get("/catalog", h.Catalog)
		r.Post("/purchase", h.Purchase)
		r.Get("/orders", h.Orders)
		r.Get("/vpn-status", h.VPNStatus)
		// eSIM simplified endpoints (destinations/plans/order)
		r.Get("/esim/destinations", h.ESIMDestinations)
		r.Get("/esim/plans", h.ESIMPlans)
		r.Post("/esim/order", h.ESIMOrder)
	})
}

func (h *Handler) RegisterPublic(r chi.Router) {
	r.Get("/sub/{ref}", h.Subscription)
}

func (h *Handler) Catalog(w http.ResponseWriter, r *http.Request) {
	filter := ports.StoreProductFilter{
		CategorySlug: r.URL.Query().Get("category"),
		Country:      r.URL.Query().Get("country"),
		Search:       r.URL.Query().Get("search"),
	}
	res, err := h.uc.Catalog(r.Context(), filter)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) Purchase(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}
	var req struct {
		ProductID string `json:"productId"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ProductID == "" {
		http.Error(w, "Invalid productId", http.StatusBadRequest)

		return
	}
	pid, err := uuid.Parse(req.ProductID)
	if err != nil {
		http.Error(w, "Invalid productId", http.StatusBadRequest)

		return
	}
	out, err := h.uc.Purchase(r.Context(), userID, pid)
	if err != nil {
		// Minimal error mapping for existing frontend UX
		if errorsIs(err, domain.ErrInsufficientFunds) {
			handler.WriteJSON(w, http.StatusPaymentRequired, map[string]string{"error": "Недостаточно средств", "code": "INSUFFICIENT_FUNDS"})

			return
		}
		handler.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error(), "code": "PURCHASE_FAILED"})

		return
	}
	handler.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}
	limit := 20
	if s := r.URL.Query().Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			limit = v
		}
	}
	list, err := h.uc.Orders(r.Context(), userID, limit)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{"orders": list})
}

func (h *Handler) VPNStatus(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		http.Error(w, "ref parameter required", http.StatusBadRequest)

		return
	}
	meta, err := h.storeRepo.GetLatestCompletedOrderMetaByProviderRef(r.Context(), ref, &userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusNotFound)

		return
	}
	// For MVP: no live traffic yet
	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"ref":           ref,
		"status":        "active",
		"upload":        0,
		"download":      0,
		"used":          0,
		"total":         0,
		"remaining":     0,
		"exhausted":     false,
		"expire_ms":     0,
		"duration_days": 0,
		"used_percent":  0,
		"meta":          meta,
	})
}

func (h *Handler) ESIMDestinations(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.ESIMDestinations(r.Context())
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{"destinations": list})
}

func (h *Handler) ESIMPlans(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	if country == "" {
		http.Error(w, "country parameter required", http.StatusBadRequest)

		return
	}

	plans, err := h.uc.ESIMPlans(r.Context(), country)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{"plans": plans})
}

func (h *Handler) ESIMOrder(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	var req struct {
		PlanID string `json:"planId"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.PlanID == "" {
		http.Error(w, "Invalid planId", http.StatusBadRequest)

		return
	}

	res, err := h.uc.ESIMOrder(r.Context(), req.PlanID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, res)
}

// Subscription is a public subscription endpoint used by VPN apps.
func (h *Handler) Subscription(w http.ResponseWriter, r *http.Request) {
	ref := chi.URLParam(r, "ref")
	if ref == "" {
		http.Error(w, "missing subscription ref", http.StatusBadRequest)

		return
	}
	o, err := h.storeRepo.GetLatestCompletedOrderByProviderRef(r.Context(), ref)
	if err != nil || o.ActivationKey == "" {
		http.Error(w, "subscription not found", http.StatusNotFound)

		return
	}
	w.Header().Set("Subscription-Userinfo", "upload=0; download=0; total=0; expire=0")
	encoded := base64.StdEncoding.EncodeToString([]byte(o.ActivationKey))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//nolint:gosec // subscription is plain text base64.
	_, _ = w.Write([]byte(encoded))
}

func errorsIs(err, target error) bool {
	return errors.Is(err, target)
}
