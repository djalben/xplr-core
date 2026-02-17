package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/djalben/xplr-core/backend/repository"
)

// AdminStatsHandler - GET /api/v1/admin/stats
func AdminStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := repository.GetAdminStats()
	if err != nil {
		log.Printf("Error fetching admin stats: %v", err)
		http.Error(w, "Failed to fetch admin stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// AdminUsersHandler - GET /api/v1/admin/users
func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := repository.GetAllUsersForAdmin()
	if err != nil {
		log.Printf("Error fetching admin users: %v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetAdminTransactionReportHandler - Выдает отчет обо всех транзакциях платформы.
func GetAdminTransactionReportHandler(w http.ResponseWriter, r *http.Request) {
	// В этом обработчике мы уверены, что пользователь является Администратором

	report, err := repository.GetAdminTransactionReport()
	if err != nil {
		log.Printf("Error fetching admin report: %v", err)
		http.Error(w, "Ошибка сервера при получении отчета", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("Error encoding admin report JSON: %v", err)
	}
}
