package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	go notifyUsersAboutNews(req.Title, req.Content, req.ImageURL)

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
// Uses NotifyUserNews for image-first layout (Telegram sendPhoto + Email image block).
func notifyUsersAboutNews(title, content, imageURL string) {
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

	// Telegram caption (max 1024 chars for sendPhoto)
	tgCaption := fmt.Sprintf("📰 <b>%s</b>\n\n%s\n\n<a href=\"https://xplr.pro/news\">Читать полностью →</a>", title, preview)

	// Email body (text part, image is prepended by NotifyUserNews)
	emailBody := fmt.Sprintf(`
    <p style="color:#cbd5e1;font-size:16px;line-height:1.5;margin:0 0 16px;font-weight:700;">📰 %s</p>
    <p style="color:#94a3b8;font-size:14px;line-height:1.7;margin:0 0 24px;white-space:pre-wrap;">%s</p>
    <div style="text-align:center;">
      <a href="https://xplr.pro/news" style="display:inline-block;padding:14px 40px;background:linear-gradient(135deg,#3b82f6,#8b5cf6);color:#fff;text-decoration:none;border-radius:12px;font-size:14px;font-weight:600;">Читать полностью</a>
    </div>`, title, preview)

	count := 0
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		go service.NotifyUserNews(userID, "Новость XPLR", tgCaption, emailBody, imageURL)
		count++
	}

	log.Printf("[NEWS-NOTIFY] 📩 Sent news notifications to %d users (image=%v)", count, imageURL != "")
}

// ── POST /api/v1/admin/upload-image — upload image to Supabase Storage ──
func AdminUploadImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[UPLOAD] 📥 Incoming upload request: method=%s content-type=%s content-length=%d",
		r.Method, r.Header.Get("Content-Type"), r.ContentLength)

	// Parse multipart form (max 5MB)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		log.Printf("[UPLOAD] ❌ ParseMultipartForm failed: %v", err)
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("[UPLOAD] ❌ FormFile('image') failed: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read image file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("[UPLOAD] 📄 File received: name=%s size=%d", header.Filename, header.Size)

	// Validate content type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	contentType := ""
	switch ext {
	case ".webp":
		contentType = "image/webp"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	default:
		log.Printf("[UPLOAD] ❌ Invalid extension: %s", ext)
		http.Error(w, "Only webp, jpg, png allowed", http.StatusBadRequest)
		return
	}

	// Read file bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[UPLOAD] ❌ io.ReadAll failed: %v", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	log.Printf("[UPLOAD] 📦 Read %d bytes, ext=%s, content-type=%s", len(fileBytes), ext, contentType)

	// Generate unique filename
	filename := fmt.Sprintf("news_%d%s", time.Now().UnixMilli(), ext)

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")
	if supabaseKey == "" {
		supabaseKey = os.Getenv("SUPABASE_ANON_KEY")
	}

	if supabaseURL == "" || supabaseKey == "" {
		log.Printf("[UPLOAD] ❌ Missing env: SUPABASE_URL=%q, SUPABASE_SERVICE_KEY set=%v, SUPABASE_ANON_KEY set=%v",
			supabaseURL, os.Getenv("SUPABASE_SERVICE_KEY") != "", os.Getenv("SUPABASE_ANON_KEY") != "")
		http.Error(w, "Storage not configured — check SUPABASE_URL and SUPABASE_SERVICE_KEY env vars", http.StatusInternalServerError)
		return
	}

	bucket := "news-images"
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucket, filename)
	log.Printf("[UPLOAD] 🚀 Uploading to Supabase: %s", uploadURL)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		log.Printf("[UPLOAD] ❌ NewRequest failed: %v", err)
		http.Error(w, "Failed to create upload request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[UPLOAD] ❌ Supabase HTTP request failed: %v", err)
		http.Error(w, "Upload request failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[UPLOAD] 📡 Supabase response: status=%d body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		http.Error(w, fmt.Sprintf("Supabase Storage error %d: %s", resp.StatusCode, string(respBody)), http.StatusInternalServerError)
		return
	}

	// Build public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucket, filename)
	log.Printf("[UPLOAD] ✅ Image uploaded successfully: %s (%d bytes)", publicURL, len(fileBytes))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":      publicURL,
		"filename": filename,
	})
}
