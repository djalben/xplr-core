package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/aalabin/xplr/middleware"
	"github.com/aalabin/xplr/models"
	"github.com/aalabin/xplr/repository"
)

// CreateTeamHandler - POST /api/v1/user/teams
func CreateTeamHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	team, err := repository.CreateTeam(userID, req.Name)
	if err != nil {
		log.Printf("Error creating team: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

// GetUserTeamsHandler - GET /api/v1/user/teams
func GetUserTeamsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teams, err := repository.GetUserTeams(userID)
	if err != nil {
		log.Printf("Error fetching teams for user %d: %v", userID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

// GetTeamHandler - GET /api/v1/user/teams/{id}
func GetTeamHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["id"]
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil || teamID <= 0 {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Проверить доступ
	hasAccess, _, err := repository.CheckTeamAccess(teamID, userID)
	if err != nil {
		log.Printf("Error checking team access: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Получить команду
	team, err := repository.GetTeam(teamID)
	if err != nil {
		log.Printf("Error fetching team %d: %v", teamID, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Получить участников
	members, err := repository.GetTeamMembers(teamID)
	if err != nil {
		log.Printf("Error fetching team members: %v", err)
		// Не возвращаем ошибку, просто пустой список участников
		members = []models.TeamMember{}
	}

	response := map[string]interface{}{
		"team":    team,
		"members": members,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// InviteTeamMemberHandler - POST /api/v1/user/teams/{id}/members
func InviteTeamMemberHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["id"]
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil || teamID <= 0 {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var req models.InviteTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "member" // По умолчанию
	}

	err = repository.InviteTeamMember(teamID, userID, req.Email, req.Role)
	if err != nil {
		if err.Error() == "access denied" || err.Error() == "insufficient permissions" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		log.Printf("Error inviting team member: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Team member invited successfully",
	})
}

// RemoveTeamMemberHandler - DELETE /api/v1/user/teams/{id}/members/{userId}
func RemoveTeamMemberHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["id"]
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil || teamID <= 0 {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	memberIDStr := vars["userId"]
	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil || memberID <= 0 {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	err = repository.RemoveTeamMember(teamID, memberID, userID)
	if err != nil {
		if err.Error() == "access denied" || err.Error() == "insufficient permissions" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		log.Printf("Error removing team member: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Team member removed successfully",
	})
}

// UpdateTeamMemberRoleHandler - PATCH /api/v1/user/teams/{id}/members/{userId}/role
func UpdateTeamMemberRoleHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["id"]
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil || teamID <= 0 {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	memberIDStr := vars["userId"]
	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil || memberID <= 0 {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateTeamMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = repository.UpdateTeamMemberRole(teamID, memberID, req.Role, userID)
	if err != nil {
		if err.Error() == "access denied" || err.Error() == "insufficient permissions" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		log.Printf("Error updating team member role: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Team member role updated successfully",
	})
}
