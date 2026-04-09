package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/gorilla/mux"
)

// ── GET /api/v1/user/news — paginated news list ──
func GetNewsHandler(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := GlobalDB.Query(`SELECT id, title, content, COALESCE(image_url, ''), created_at FROM news ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch news", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type NewsItem struct {
		ID        int       `json:"id"`
		Title     string    `json:"title"`
		Content   string    `json:"content"`
		ImageURL  string    `json:"image_url"`
		CreatedAt time.Time `json:"created_at"`
	}

	var news []NewsItem
	for rows.Next() {
		var n NewsItem
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.ImageURL, &n.CreatedAt); err != nil {
			continue
		}
		news = append(news, n)
	}
	if news == nil {
		news = []NewsItem{}
	}

	var total int
	GlobalDB.QueryRow(`SELECT COUNT(*) FROM news`).Scan(&total)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": news,
		"total": total,
	})
}

// ── GET /api/v1/user/news-notifications — get user's news notification preference ──
func GetNewsNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var enabled bool
	err := GlobalDB.QueryRow(`SELECT COALESCE(news_notifications_enabled, TRUE) FROM users WHERE id = $1`, userID).Scan(&enabled)
	if err != nil {
		http.Error(w, "Failed to get preference", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": enabled})
}

// ── PATCH /api/v1/user/news-notifications — toggle news notifications ──
func UpdateNewsNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	_, err := GlobalDB.Exec(`UPDATE users SET news_notifications_enabled = $1 WHERE id = $2`, req.Enabled, userID)
	if err != nil {
		http.Error(w, "Failed to update preference", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ── POST /api/v1/admin/news — create news + notify users ──
func AdminCreateNewsHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)

	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	var newsID int
	err := GlobalDB.QueryRow(`INSERT INTO news (title, content, image_url) VALUES ($1, $2, $3) RETURNING id`,
		req.Title, req.Content, req.ImageURL).Scan(&newsID)
	if err != nil {
		log.Printf("[NEWS] ❌ Failed to create news: %v", err)
		http.Error(w, "Failed to create news", http.StatusInternalServerError)
		return
	}

	repository.WriteAdminLog(adminID, fmt.Sprintf("Создана новость #%d: %s", newsID, req.Title))
	log.Printf("[NEWS] ✅ Admin %d created news #%d: %s", adminID, newsID, req.Title)

	// Notify users with news_notifications_enabled in background
	go notifyUsersAboutNews(req.Title, req.Content)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    newsID,
		"title": req.Title,
	})
}

// ── DELETE /api/v1/admin/news/{id} — delete news article ──
func AdminDeleteNewsHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	newsID, err := strconv.Atoi(vars["id"])
	if err != nil || newsID <= 0 {
		http.Error(w, "Invalid news id", http.StatusBadRequest)
		return
	}

	res, err := GlobalDB.Exec(`DELETE FROM news WHERE id = $1`, newsID)
	if err != nil {
		http.Error(w, "Failed to delete news", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "News not found", http.StatusNotFound)
		return
	}

	repository.WriteAdminLog(adminID, fmt.Sprintf("Удалена новость #%d", newsID))
	log.Printf("[NEWS] 🗑️ Admin %d deleted news #%d", adminID, newsID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// notifyUsersAboutNews sends news notification to all opted-in users.
func notifyUsersAboutNews(title, content string) {
	if GlobalDB == nil {
		return
	}

	rows, err := GlobalDB.Query(`SELECT id FROM users WHERE COALESCE(news_notifications_enabled, TRUE) = TRUE`)
	if err != nil {
		log.Printf("[NEWS-NOTIFY] ❌ Failed to query users: %v", err)
		return
	}
	defer rows.Close()

	// Truncate content for notification preview
	preview := content
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}

	msg := fmt.Sprintf("📰 <b>%s</b>\n\n%s\n\n<a href=\"https://xplr.pro/news\">Читать полностью</a>", title, preview)

	count := 0
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		go service.NotifyUser(userID, "Новость XPLR", msg)
		count++
	}

	log.Printf("[NEWS-NOTIFY] 📩 Sent news notifications to %d users", count)
}
