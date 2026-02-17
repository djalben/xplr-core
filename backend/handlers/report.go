package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
)

// GetUserTransactionReportHandler обрабатывает запрос GET /api/v1/user/report
// Поддерживает query параметры: start_date, end_date, transaction_type, status, card_id, search, limit, offset
func GetUserTransactionReportHandler(w http.ResponseWriter, r *http.Request) {
	// AuthMiddleware гарантирует наличие UserID в контексте
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		log.Println("Error: User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Парсим query параметры
	filters := make(map[string]interface{})
	query := r.URL.Query()

	if startDate := query.Get("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := query.Get("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}
	if transactionType := query.Get("transaction_type"); transactionType != "" {
		filters["transaction_type"] = transactionType
	}
	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}
	if cardIDStr := query.Get("card_id"); cardIDStr != "" {
		if cardID, err := strconv.Atoi(cardIDStr); err == nil && cardID > 0 {
			filters["card_id"] = cardID
		}
	}
	if search := query.Get("search"); search != "" {
		filters["search"] = search
	}
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters["limit"] = limit
		}
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters["offset"] = offset
		}
	}

	report, err := repository.GetUserTransactionReport(userID, filters)
	if err != nil {
		log.Printf("Failed to fetch report for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("Failed to encode response for user %d report: %v", userID, err)
	}
}

// GetAdminTransactionReportHandler удален, чтобы избежать конфликта с handlers/admin.go
