package models

import "time"

// ReportSummary - Агрегированная сводка по кликам
// Это главная структура, которую мы будем возвращать пользователю (UI-отчет).
type ReportSummary struct {
    Date             time.Time `json:"date"`               // Дата, по которой сгруппированы данные
    SubID            string    `json:"sub_id"`             // SubID, по которому сгруппированы данные
    TotalClicks      int       `json:"total_clicks"`       // Общее количество кликов за период
    UniqueClicks     int       `json:"unique_clicks"`      // Количество уникальных кликов (по IP)
    // Spend - Будет добавлено, когда реализуем финансовую логику (ЭТАП 2)
    // Conversions - Будет добавлено, когда реализуем логику конверсий
}

// SpendReportSummary - Агрегированная сводка по расходам (для трекеров)
// Используется для API-интеграции (GET /api/v1/data/spends)
type SpendReportSummary struct {
    Date             time.Time `json:"date"`               // Дата, по которой сгруппированы данные
    SubID            string    `json:"sub_id"`             // SubID, по которому сгруппированы данные
    TotalSpend       float64   `json:"total_spend"`        // Общая сумма трат за период (в USD)
    TransactionCount int       `json:"transaction_count"`  // Общее количество транзакций
}

// ReportRequest - Структура для парсинга параметров запроса отчета
type ReportRequest struct {
    StartDate time.Time `json:"start_date"`
    EndDate   time.Time `json:"end_date"`
    SubID     string    `json:"sub_id"` // Опциональный фильтр
    GroupBy   string    `json:"group_by"` // 'day', 'subid', 'offer'
}