package admin

import (
	"net/http"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterStore(r chi.Router) {
	r.Get("/store/products", h.AdminStoreProducts)
	r.Patch("/store/products/{id}", h.AdminPatchStoreProduct)
	r.Post("/store/bulk-markup", h.AdminBulkMarkup)
}

func (h *Handler) AdminStoreProducts(w http.ResponseWriter, r *http.Request) {
	var f ports.StoreAdminProductFilter

	if t := r.URL.Query().Get("product_type"); t != "" {
		pt := domain.StoreProductType(t)
		f.ProductType = &pt
	}

	list, err := h.storeRepo.AdminListProducts(r.Context(), f)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, list)
}

func (h *Handler) AdminPatchStoreProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := adminChiUUID(w, r)
	if !ok {
		return
	}

	var req struct {
		CostPrice      *float64 `json:"cost_price"`
		MarkupPercent  *float64 `json:"markup_percent"`
		ImageURL       *string  `json:"image_url"`
		RetailPriceUSD *float64 `json:"retail_price"`
		InStock        *bool    `json:"in_stock"`
		Meta           *string  `json:"meta"`
		SortOrder      *int     `json:"sort_order"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	p, err := h.storeRepo.GetProductByID(r.Context(), id)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusNotFound)

		return
	}

	if req.CostPrice != nil {
		p.CostPrice = domain.NewNumeric(*req.CostPrice)
	}
	if req.MarkupPercent != nil {
		p.MarkupPct = domain.NewNumeric(*req.MarkupPercent)
	}
	if req.ImageURL != nil {
		p.ImageURL = *req.ImageURL
	}
	if req.RetailPriceUSD != nil {
		p.PriceUSD = domain.NewNumeric(*req.RetailPriceUSD)
	}
	if req.InStock != nil {
		p.InStock = *req.InStock
	}
	if req.Meta != nil {
		p.Meta = *req.Meta
	}
	if req.SortOrder != nil {
		p.SortOrder = *req.SortOrder
	}

	_ = time.Now()

	err = h.storeRepo.AdminUpdateProduct(r.Context(), p)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) AdminBulkMarkup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Delta       float64 `json:"delta"`
		ProductType string  `json:"product_type"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	pt := domain.StoreProductType(req.ProductType)
	switch pt {
	case domain.StoreProductTypeESIM, domain.StoreProductTypeDigital, domain.StoreProductTypeVPN:
	default:
		http.Error(w, "product_type must be esim, digital or vpn", http.StatusBadRequest)

		return
	}

	affected, err := h.storeRepo.AdminBulkAddMarkup(r.Context(), pt, domain.NewNumeric(req.Delta))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]any{"affected": affected})
}
