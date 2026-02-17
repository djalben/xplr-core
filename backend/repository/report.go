package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/shopspring/decimal"
)

// ReportSummary - Структура для агрегированного отчета по транзакциям.
type ReportSummary struct {
	TotalTransactions int                  `json:"total_transactions"`
	TotalAmount       decimal.Decimal      `json:"total_amount"`
	TotalFee          decimal.Decimal      `json:"total_fee"`
	Transactions      []models.Transaction `json:"transactions"`
}

// GetUserTransactionReport - Извлекает отчет по транзакциям только для данного пользователя.
// Поддерживает фильтры: startDate, endDate, transactionType, status, cardId
func GetUserTransactionReport(userID int, filters map[string]interface{}) (ReportSummary, error) {
	if GlobalDB == nil {
		return ReportSummary{}, fmt.Errorf("database connection not initialized")
	}

	// Базовый запрос
	queryTx := `
        SELECT id, user_id, card_id, amount, fee, transaction_type, status, details, executed_at
        FROM transactions
        WHERE user_id = $1
    `
	args := []interface{}{userID}
	argIndex := 2

	// Применяем фильтры
	if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
		queryTx += fmt.Sprintf(" AND executed_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}
	if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
		queryTx += fmt.Sprintf(" AND executed_at <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}
	if transactionType, ok := filters["transaction_type"].(string); ok && transactionType != "" {
		queryTx += fmt.Sprintf(" AND transaction_type = $%d", argIndex)
		args = append(args, transactionType)
		argIndex++
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		queryTx += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	if cardID, ok := filters["card_id"].(int); ok && cardID > 0 {
		queryTx += fmt.Sprintf(" AND card_id = $%d", argIndex)
		args = append(args, cardID)
		argIndex++
	}

	// Поиск по details (если указан)
	if search, ok := filters["search"].(string); ok && search != "" {
		queryTx += fmt.Sprintf(" AND details ILIKE $%d", argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	queryTx += " ORDER BY executed_at DESC"

	// Лимит результатов (по умолчанию 100, максимум 1000)
	limit := 100
	if limitVal, ok := filters["limit"].(int); ok && limitVal > 0 {
		if limitVal > 1000 {
			limit = 1000
		} else {
			limit = limitVal
		}
	}
	queryTx += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)
	argIndex++

	// Offset для пагинации
	if offset, ok := filters["offset"].(int); ok && offset > 0 {
		queryTx += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	rows, err := GlobalDB.Query(queryTx, args...)
	if err != nil {
		log.Printf("DB Error fetching user %d transactions: %v", userID, err)
		return ReportSummary{}, fmt.Errorf("ошибка сервера при получении транзакций")
	}
	defer rows.Close()

	var report ReportSummary
	report.Transactions = []models.Transaction{}
	report.TotalAmount = decimal.Zero
	report.TotalFee = decimal.Zero

	for rows.Next() {
		var tx models.Transaction
		
		err := rows.Scan(
			&tx.TransactionID, 
			&tx.UserID, 
			&tx.Amount, 
			&tx.Fee, 
			&tx.TransactionType, 
			&tx.Status, 
			&tx.Details, 
			&tx.ExecutedAt,
		)
		if err != nil {
			log.Printf("DB Error scanning transaction: %v", err)
			continue 
		}
		
		report.Transactions = append(report.Transactions, tx)
		report.TotalTransactions++
		report.TotalAmount = report.TotalAmount.Add(tx.Amount)
		report.TotalFee = report.TotalFee.Add(tx.Fee) 
	}

	if err := rows.Err(); err != nil {
		return ReportSummary{}, fmt.Errorf("ошибка итерации по результатам транзакций: %w", err)
	}

	// Decimal.Decimal уже поддерживает точность, округление не требуется
	return report, nil
}

// GetAdminTransactionReport - Извлекает отчет по всем транзакциям (для админа).
func GetAdminTransactionReport() (ReportSummary, error) {
    if GlobalDB == nil {
		return ReportSummary{}, fmt.Errorf("database connection not initialized")
	}
    
	queryTx := `
        SELECT id, user_id, amount, fee, transaction_type, status, details, executed_at
        FROM transactions
        ORDER BY executed_at DESC
    `
	rows, err := GlobalDB.Query(queryTx)
	if err != nil {
		log.Printf("DB Error fetching admin transactions: %v", err)
		return ReportSummary{}, fmt.Errorf("ошибка сервера при получении всех транзакций")
	}
	defer rows.Close()

	var report ReportSummary
	report.Transactions = []models.Transaction{}
	report.TotalAmount = decimal.Zero
	report.TotalFee = decimal.Zero

	for rows.Next() {
        var tx models.Transaction
		var cardID sql.NullInt64
		
		err := rows.Scan(
			&tx.TransactionID, 
			&tx.UserID, 
			&cardID,
			&tx.Amount, 
			&tx.Fee, 
			&tx.TransactionType, 
			&tx.Status, 
			&tx.Details, 
			&tx.ExecutedAt,
		)
		if cardID.Valid {
			cardIDVal := int(cardID.Int64)
			tx.CardID = &cardIDVal
		}
		if err != nil {
			log.Printf("DB Error scanning admin transaction: %v", err)
			continue 
		}
		
		report.Transactions = append(report.Transactions, tx)
		report.TotalTransactions++
		report.TotalAmount = report.TotalAmount.Add(tx.Amount)
		report.TotalFee = report.TotalFee.Add(tx.Fee)
	}

	if err := rows.Err(); err != nil {
		return ReportSummary{}, fmt.Errorf("ошибка итерации по результатам транзакций (админ): %w", err)
	}

	// Decimal.Decimal уже поддерживает точность, округление не требуется
	return report, nil
}