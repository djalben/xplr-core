package news

import (
	"net/http"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	newsRepo ports.NewsRepository
	userRepo ports.UserRepository
}

func NewHandler(newsRepo ports.NewsRepository, userRepo ports.UserRepository) *Handler {
	return &Handler{newsRepo: newsRepo, userRepo: userRepo}
}

func (h *Handler) Register(r chi.Router) {
	// NOTE: user handler already mounts `Route("/user", ...)`.
	// Register news endpoints as absolute paths under `/v1` to avoid double-mounting `/user`.
	r.Get("/user/news", h.List)
	r.Get("/user/news/unread-count", h.UnreadCount)
	r.Post("/user/news/mark-as-read", h.MarkAsRead)
	r.Get("/user/news-notifications", h.GetNotifications)
	r.Patch("/user/news-notifications", h.PatchNotifications)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0

	if s := r.URL.Query().Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			limit = v
		}
	}
	if s := r.URL.Query().Get("offset"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			offset = v
		}
	}

	items, total, err := h.newsRepo.ListPublished(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	u, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	if u.LastReadNewsAt == nil {
		_, total, err := h.newsRepo.ListPublished(r.Context(), 1, 0)
		if err != nil {
			http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

			return
		}
		handler.WriteJSON(w, http.StatusOK, map[string]any{"count": total})

		return
	}

	// MVP: without a dedicated repository method we can't do an efficient DB count here.
	// Return 0 once the user has marked news as read; upgrade later when we introduce ListPublishedSince/CountPublishedSince.
	handler.WriteJSON(w, http.StatusOK, map[string]any{"count": 0})
}

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	u, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	now := time.Now().UTC()
	u.LastReadNewsAt = &now

	err = h.userRepo.Update(r.Context(), u)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	u, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{"enabled": u.NewsNotificationsEnabled})
}

func (h *Handler) PatchNotifications(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	u, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	u.NewsNotificationsEnabled = req.Enabled

	err = h.userRepo.Update(r.Context(), u)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

var _ = domain.NewsArticle{}
