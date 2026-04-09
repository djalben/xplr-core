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

// getSupabaseConfig reads and validates Supabase env vars.
// Returns url, key, keySource. Empty url/key means not configured.
func getSupabaseConfig() (string, string, string) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")
	keySource := "SUPABASE_SERVICE_KEY"
	if supabaseKey == "" {
		supabaseKey = os.Getenv("SUPABASE_ANON_KEY")
		keySource = "SUPABASE_ANON_KEY"
	}
	if supabaseKey == "" {
		keySource = "NONE"
	}
	return supabaseURL, supabaseKey, keySource
}

// safePrefix returns the first n chars of a string for safe logging.
func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// checkSupabaseBucket verifies the bucket exists by listing it.
func checkSupabaseBucket(supabaseURL, supabaseKey, bucket string) (bool, string) {
	checkURL := fmt.Sprintf("%s/storage/v1/bucket/%s", supabaseURL, bucket)
	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return false, fmt.Sprintf("NewRequest error: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("apikey", supabaseKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("HTTP error: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		return true, string(body)
	}
	return false, fmt.Sprintf("status=%d body=%s", resp.StatusCode, string(body))
}

// ── GET /api/v1/admin/test-upload — test Supabase Storage connectivity ──
func AdminTestUploadHandler(w http.ResponseWriter, r *http.Request) {
	supabaseURL, supabaseKey, keySource := getSupabaseConfig()

	result := map[string]interface{}{
		"supabase_url": supabaseURL,
		"key_source":   keySource,
		"key_prefix":   safePrefix(supabaseKey, 8),
		"key_length":   len(supabaseKey),
		"url_set":      supabaseURL != "",
		"key_set":      supabaseKey != "",
	}

	if supabaseURL == "" || supabaseKey == "" {
		result["error"] = "SUPABASE_URL or key not configured"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result)
		return
	}

	// Check bucket existence
	bucketOK, bucketInfo := checkSupabaseBucket(supabaseURL, supabaseKey, "news-images")
	result["bucket_exists"] = bucketOK
	result["bucket_info"] = bucketInfo

	if !bucketOK {
		log.Printf("[TEST-UPLOAD] ❌ BUCKET NOT FOUND: news-images — %s", bucketInfo)
	}

	// Try uploading a tiny test file
	testFilename := fmt.Sprintf("_test_%d.txt", time.Now().UnixMilli())
	uploadURL := fmt.Sprintf("%s/storage/v1/object/news-images/%s", supabaseURL, testFilename)
	testBody := []byte("test-upload-connectivity-check")
	req, _ := http.NewRequest("POST", uploadURL, bytes.NewReader(testBody))
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("x-upsert", "true")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result["upload_test"] = fmt.Sprintf("HTTP error: %v", err)
	} else {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		result["upload_test_status"] = resp.StatusCode
		result["upload_test_body"] = string(respBody)

		if resp.StatusCode < 400 {
			result["upload_test"] = "SUCCESS"
			// Clean up test file
			delURL := fmt.Sprintf("%s/storage/v1/object/news-images/%s", supabaseURL, testFilename)
			delReq, _ := http.NewRequest("DELETE", delURL, nil)
			delReq.Header.Set("Authorization", "Bearer "+supabaseKey)
			delReq.Header.Set("apikey", supabaseKey)
			client.Do(delReq)
		} else {
			result["upload_test"] = "FAILED"
		}
	}

	log.Printf("[TEST-UPLOAD] Result: %+v", result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ── POST /api/v1/admin/upload-image — upload image to Supabase Storage ──
func AdminUploadImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[UPLOAD] ═══════════════════════════════════════════")
	log.Printf("[UPLOAD] 📥 Request: method=%s content-type=%q content-length=%d",
		r.Method, r.Header.Get("Content-Type"), r.ContentLength)

	// ── Step 1: Validate env vars ──
	supabaseURL, supabaseKey, keySource := getSupabaseConfig()
	log.Printf("[UPLOAD] 🔑 ENV: SUPABASE_URL=%q, key_source=%s, key_prefix=%s, key_len=%d",
		supabaseURL, keySource, safePrefix(supabaseKey, 8), len(supabaseKey))

	if supabaseURL == "" || supabaseKey == "" {
		errMsg := fmt.Sprintf("Storage not configured: SUPABASE_URL=%q, key_source=%s", supabaseURL, keySource)
		log.Printf("[UPLOAD] ❌ %s", errMsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
		return
	}

	// ── Step 2: Check bucket exists ──
	bucketOK, bucketInfo := checkSupabaseBucket(supabaseURL, supabaseKey, "news-images")
	log.Printf("[UPLOAD] 🪣 Bucket check: exists=%v info=%s", bucketOK, bucketInfo)
	if !bucketOK {
		log.Printf("[UPLOAD] ❌ BUCKET NOT FOUND: news-images")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":       "BUCKET NOT FOUND: news-images",
			"bucket_info": bucketInfo,
			"hint":        "Create bucket 'news-images' in Supabase Dashboard → Storage → New Bucket (public=true)",
		})
		return
	}

	// ── Step 3: Parse multipart form ──
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		log.Printf("[UPLOAD] ❌ ParseMultipartForm failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("ParseMultipartForm: %v", err)})
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("[UPLOAD] ❌ FormFile('image') failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("FormFile: %v", err)})
		return
	}
	defer file.Close()

	log.Printf("[UPLOAD] 📄 File: name=%q size=%d", header.Filename, header.Size)

	// ── Step 4: Validate file type ──
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
		log.Printf("[UPLOAD] ❌ Invalid extension: %q", ext)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Invalid extension: %s", ext)})
		return
	}

	// ── Step 5: Read file bytes ──
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[UPLOAD] ❌ ReadAll failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("ReadAll: %v", err)})
		return
	}
	log.Printf("[UPLOAD] 📦 Read %d bytes, content-type=%s", len(fileBytes), contentType)

	if len(fileBytes) == 0 {
		log.Printf("[UPLOAD] ❌ Empty file")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "File is empty"})
		return
	}

	// ── Step 6: Upload to Supabase Storage ──
	filename := fmt.Sprintf("news_%d%s", time.Now().UnixMilli(), ext)
	uploadURL := fmt.Sprintf("%s/storage/v1/object/news-images/%s", supabaseURL, filename)
	log.Printf("[UPLOAD] 🚀 Uploading: %s (%d bytes, %s)", uploadURL, len(fileBytes), contentType)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		log.Printf("[UPLOAD] ❌ NewRequest failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("NewRequest: %v", err)})
		return
	}
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[UPLOAD] ❌ Supabase HTTP error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Supabase HTTP: %v", err)})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[UPLOAD] 📡 Supabase response: status=%d body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":           fmt.Sprintf("Supabase Storage error %d", resp.StatusCode),
			"supabase_body":   string(respBody),
			"upload_url":      uploadURL,
			"content_type":    contentType,
			"file_size_bytes": fmt.Sprintf("%d", len(fileBytes)),
		})
		return
	}

	// ── Step 7: Return public URL ──
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/news-images/%s", supabaseURL, filename)
	log.Printf("[UPLOAD] ✅ SUCCESS: %s (%d bytes)", publicURL, len(fileBytes))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":      publicURL,
		"filename": filename,
	})
}
