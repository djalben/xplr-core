package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aalabin/xplr/middleware"
	"github.com/aalabin/xplr/repository"
)

// GetUserGradeHandler - GET /api/v1/user/grade
func GetUserGradeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	gradeInfo, err := repository.GetUserGradeInfo(userID)
	if err != nil {
		log.Printf("Error fetching grade info for user %d: %v", userID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gradeInfo)
}
