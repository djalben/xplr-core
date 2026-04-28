package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func (h *Handler) RegisterNews(r chi.Router) {
	r.Get("/news", h.AdminListNews)
	r.Post("/news", h.AdminCreateNews)
	r.Put("/news/{id}", h.AdminUpdateNews)
	r.Patch("/news/{id}", h.AdminPatchNews)
	r.Delete("/news/{id}", h.AdminDeleteNews)
}

// AdminListNews — GET /admin/news?limit= (MVP: без offset/status-фильтра, берём все).
func (h *Handler) AdminListNews(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if s := r.URL.Query().Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			limit = v
		}
	}

	list, err := h.newsRepo.ListAll(r.Context(), limit)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, list)
}

// AdminCreateNews — POST /admin/news.
func (h *Handler) AdminCreateNews(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
		Status   string `json:"status"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}
	if req.Title == "" || req.Content == "" {
		http.Error(w, "title and content are required", http.StatusBadRequest)

		return
	}

	st := domain.NewsDraft
	if req.Status != "" {
		switch domain.NewsStatus(req.Status) {
		case domain.NewsDraft, domain.NewsPublished, domain.NewsArchived:
			st = domain.NewsStatus(req.Status)
		default:
			http.Error(w, "invalid status", http.StatusBadRequest)

			return
		}
	}

	now := time.Now().UTC()
	a := &domain.NewsArticle{
		ID:        domain.NewUUID(),
		Title:     req.Title,
		Content:   req.Content,
		ImageURL:  req.ImageURL,
		Status:    st,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := h.newsRepo.Create(r.Context(), a)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, a)
}

// AdminUpdateNews — PUT /admin/news/{id}.
func (h *Handler) AdminUpdateNews(w http.ResponseWriter, r *http.Request) {
	id, ok := adminChiUUID(w, r)
	if !ok {
		return
	}

	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
		Status   string `json:"status"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	if req.Title == "" || req.Content == "" {
		http.Error(w, "title and content are required", http.StatusBadRequest)

		return
	}

	a, err := h.newsRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusNotFound)

		return
	}

	a.Title = req.Title
	a.Content = req.Content
	a.ImageURL = req.ImageURL

	if req.Status != "" {
		switch domain.NewsStatus(req.Status) {
		case domain.NewsDraft, domain.NewsPublished, domain.NewsArchived:
			a.Status = domain.NewsStatus(req.Status)
		default:
			http.Error(w, "invalid status", http.StatusBadRequest)

			return
		}
	}

	a.UpdatedAt = time.Now().UTC()

	err = h.newsRepo.Update(r.Context(), a)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, a)
}

// AdminPatchNews — PATCH /admin/news/{id} (MVP: только status).
func (h *Handler) AdminPatchNews(w http.ResponseWriter, r *http.Request) {
	id, ok := adminChiUUID(w, r)
	if !ok {
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}
	if req.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)

		return
	}

	st := domain.NewsStatus(req.Status)
	switch st {
	case domain.NewsDraft, domain.NewsPublished, domain.NewsArchived:
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)

		return
	}

	err := h.newsRepo.SetStatus(r.Context(), id, st)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]string{"status": "success"})
}

// AdminDeleteNews — DELETE /admin/news/{id}.
func (h *Handler) AdminDeleteNews(w http.ResponseWriter, r *http.Request) {
	id, ok := adminChiUUID(w, r)
	if !ok {
		return
	}

	err := h.newsRepo.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusNotFound)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]string{"status": "success"})
}
