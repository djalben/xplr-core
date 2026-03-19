package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
)

// GetUnifiedTransactionsHandler — GET /api/v1/user/transactions
// Единый эндпоинт для всех денежных потоков (кошелёк, карты, рефералы).
// Query params: start_date, end_date, source_type, search, limit, offset
func GetUnifiedTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filters := make(map[string]interface{})
	q := r.URL.Query()

	if v := q.Get("start_date"); v != "" {
		filters["start_date"] = v
	}
	if v := q.Get("end_date"); v != "" {
		filters["end_date"] = v
	}
	if v := q.Get("source_type"); v != "" {
		filters["source_type"] = v
	}
	if v := q.Get("card_id"); v != "" {
		if cid, err := strconv.Atoi(v); err == nil {
			if cid == 0 {
				filters["card_id_wallet"] = true
			} else if cid > 0 {
				filters["card_id"] = cid
			}
		}
	}
	if v := q.Get("search"); v != "" {
		filters["search"] = v
	}
	if v := q.Get("limit"); v != "" {
		if lim, err := strconv.Atoi(v); err == nil && lim > 0 {
			filters["limit"] = lim
		}
	}
	if v := q.Get("offset"); v != "" {
		if off, err := strconv.Atoi(v); err == nil && off >= 0 {
			filters["offset"] = off
		}
	}

	txs, total, err := repository.GetUnifiedTransactions(userID, filters)
	if err != nil {
		log.Printf("Failed to fetch transactions for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": txs,
		"total":        total,
	})
}
