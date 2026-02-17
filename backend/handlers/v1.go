package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	// Другие импорты, необходимые для остальных хендлеров
)

// GlobalDB - Переменная для подключения к БД. Объявляется только здесь в пакете handlers.
var GlobalDB *sql.DB

// HealthCheckHandler - Простой хендлер для проверки работоспособности.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if GlobalDB == nil {
		http.Error(w, "Database connection not initialized", http.StatusInternalServerError)
		return
	}

	// Проверяем соединение с БД
	err := GlobalDB.Ping()
	if err != nil {
		http.Error(w, fmt.Sprintf("Database ping failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is up and running. Database connected."))
}

// GetReportDataHandler теперь находится только в handlers/report.go
// Другие хендлеры, определенные здесь, должны быть ниже.