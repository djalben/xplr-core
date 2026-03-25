package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// GetSystemSettingsHandler - GET /api/v1/admin/system-settings
// Returns all system settings (admin only)
func GetSystemSettingsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := GlobalDB.Query(`SELECT setting_key, setting_value, description FROM system_settings ORDER BY setting_key`)
	if err != nil {
		log.Printf("[SYSTEM-SETTINGS] Query error: %v", err)
		http.Error(w, "Failed to fetch settings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var settings []map[string]string
	for rows.Next() {
		var key, value, desc string
		if err := rows.Scan(&key, &value, &desc); err != nil {
			continue
		}
		settings = append(settings, map[string]string{
			"key":         key,
			"value":       value,
			"description": desc,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSystemSettingHandler - PATCH /api/v1/admin/system-settings/{key}
// Updates a system setting value (admin only)
func UpdateSystemSettingHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Extract key from URL path
	key := r.URL.Path[len("/api/v1/admin/system-settings/"):]
	if key == "" {
		http.Error(w, "Setting key required", http.StatusBadRequest)
		return
	}

	_, err := GlobalDB.Exec(`UPDATE system_settings SET setting_value = $1, updated_at = NOW() WHERE setting_key = $2`, req.Value, key)
	if err != nil {
		log.Printf("[SYSTEM-SETTINGS] Update error: %v", err)
		http.Error(w, "Failed to update setting", http.StatusInternalServerError)
		return
	}

	log.Printf("[SYSTEM-SETTINGS] ✅ Updated %s = %s", key, req.Value)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
		"key":    key,
		"value":  req.Value,
	})
}

// GetSBPStatusHandler - GET /api/v1/sbp-status
// Public endpoint to check if SBP is enabled
func GetSBPStatusHandler(w http.ResponseWriter, r *http.Request) {
	var enabled string
	err := GlobalDB.QueryRow(`SELECT setting_value FROM system_settings WHERE setting_key = 'sbp_enabled'`).Scan(&enabled)
	if err != nil {
		enabled = "true" // default to enabled if setting doesn't exist
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sbp_enabled": enabled == "true",
	})
}
