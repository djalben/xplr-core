package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aalabin/xplr/repository"
)

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