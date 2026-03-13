package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// AdminDashboardStatsHandler - GET /api/v1/admin/dashboard
func AdminDashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := repository.GetAdminDashboardStats()
	if err != nil {
		log.Printf("AdminDashboardStats: error: %v", err)
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// AdminSearchUsersHandler - GET /api/v1/admin/users/search?q=email&limit=50
func AdminSearchUsersHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if strings.TrimSpace(q) == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	users, err := repository.SearchUsersByEmail(q, limit)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// AdminUpdateUserGradeHandler - PATCH /api/v1/admin/users/{id}/grade
func AdminUpdateUserGradeHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	targetID, err := strconv.Atoi(vars["id"])
	if err != nil || targetID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	var req struct {
		Grade string `json:"grade"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	grade := strings.TrimSpace(strings.ToUpper(req.Grade))
	if err := repository.AdminUpdateUserGrade(targetID, grade); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	repository.WriteAdminLog(adminID, fmt.Sprintf("Изменен грейд юзера %d на %s", targetID, grade))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"user_id": targetID, "grade": grade})
}

// AdminGetCommissionConfigHandler - GET /api/v1/admin/commissions
func AdminGetCommissionConfigHandler(w http.ResponseWriter, r *http.Request) {
	configs, err := repository.GetAllCommissionConfigs()
	if err != nil {
		http.Error(w, "Failed to fetch commission config", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}

// AdminUpdateCommissionConfigHandler - PATCH /api/v1/admin/commissions/{id}
func AdminUpdateCommissionConfigHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	configID, err := strconv.Atoi(vars["id"])
	if err != nil || configID <= 0 {
		http.Error(w, "invalid config id", http.StatusBadRequest)
		return
	}
	var req struct {
		Value decimal.Decimal `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := repository.UpdateCommissionConfig(configID, req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repository.WriteAdminLog(adminID, fmt.Sprintf("Изменена комиссия config_id=%d, value=%s", configID, req.Value.String()))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": configID, "value": req.Value.String()})
}

// AdminGetSupportTicketsHandler - GET /api/v1/admin/tickets?status=open
func AdminGetSupportTicketsHandler(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	tickets, err := repository.GetAllSupportTickets(statusFilter)
	if err != nil {
		http.Error(w, "Failed to fetch tickets", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

// AdminUpdateTicketStatusHandler - PATCH /api/v1/admin/tickets/{id}
func AdminUpdateTicketStatusHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	ticketID, err := strconv.Atoi(vars["id"])
	if err != nil || ticketID <= 0 {
		http.Error(w, "invalid ticket id", http.StatusBadRequest)
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := repository.UpdateSupportTicketStatus(ticketID, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repository.WriteAdminLog(adminID, fmt.Sprintf("Обновлен статус тикета %d на %s", ticketID, req.Status))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": ticketID, "status": req.Status})
}

// AdminEmergencyFreezeHandler - POST /api/v1/admin/users/{id}/emergency-freeze
func AdminEmergencyFreezeHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	targetID, err := strconv.Atoi(vars["id"])
	if err != nil || targetID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	frozenCards, err := repository.EmergencyFreezeUser(targetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repository.WriteAdminLog(adminID, fmt.Sprintf("🚨 EMERGENCY FREEZE юзера %d — %d карт заморожено, статус BANNED, баланс обнулён", targetID, frozenCards))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":      targetID,
		"frozen_cards": frozenCards,
		"status":       "BANNED",
		"balance":      "0",
	})
}

// AdminGetLogsHandler - GET /api/v1/admin/logs
func AdminGetLogsHandler(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	logs, err := repository.GetAdminLogs(limit)
	if err != nil {
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
