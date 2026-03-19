package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/shopspring/decimal"
)

// RecordTransaction — записывает единую транзакцию в таблицу transactions.
// sourceType: wallet_topup, card_transfer, card_charge, referral_bonus, refund, commission
func RecordTransaction(userID int, cardID *int, amount decimal.Decimal, fee decimal.Decimal, txType, status, details, sourceType string, sourceID *int, currency string) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}
	if currency == "" {
		currency = "USD"
	}

	var txID int
	err := GlobalDB.QueryRow(`
		INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, source_type, source_id, currency, executed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		RETURNING id
	`, userID, cardID, amount, fee, txType, status, details, sourceType, sourceID, currency).Scan(&txID)
	if err != nil {
		return 0, fmt.Errorf("failed to record transaction: %w", err)
	}
	return txID, nil
}

// GetUnifiedTransactions — получить все транзакции пользователя с JOIN на cards для last4.
// Поддерживает фильтры: start_date, end_date, source_type, search, limit, offset.
func GetUnifiedTransactions(userID int, filters map[string]interface{}) ([]models.Transaction, int, error) {
	if GlobalDB == nil {
		return nil, 0, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT t.id, t.user_id, t.card_id, t.amount, t.fee, t.transaction_type, t.status,
		       COALESCE(t.details, ''), COALESCE(t.provider_tx_id, ''),
		       COALESCE(t.source_type, 'card_charge'), t.source_id, COALESCE(t.currency, 'USD'),
		       t.executed_at,
		       COALESCE(c.last_4_digits, '')
		FROM transactions t
		LEFT JOIN cards c ON t.card_id = c.id
		WHERE t.user_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id = $1`

	args := []interface{}{userID}
	countArgs := []interface{}{userID}
	argIdx := 2
	countArgIdx := 2

	if v, ok := filters["start_date"].(string); ok && v != "" {
		query += fmt.Sprintf(" AND t.executed_at >= $%d", argIdx)
		countQuery += fmt.Sprintf(" AND executed_at >= $%d", countArgIdx)
		args = append(args, v)
		countArgs = append(countArgs, v)
		argIdx++
		countArgIdx++
	}
	if v, ok := filters["end_date"].(string); ok && v != "" {
		query += fmt.Sprintf(" AND t.executed_at <= $%d::date + interval '1 day'", argIdx)
		countQuery += fmt.Sprintf(" AND executed_at <= $%d::date + interval '1 day'", countArgIdx)
		args = append(args, v)
		countArgs = append(countArgs, v)
		argIdx++
		countArgIdx++
	}
	if v, ok := filters["source_type"].(string); ok && v != "" {
		query += fmt.Sprintf(" AND t.source_type = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND source_type = $%d", countArgIdx)
		args = append(args, v)
		countArgs = append(countArgs, v)
		argIdx++
		countArgIdx++
	}
	if v, ok := filters["card_id"].(int); ok && v > 0 {
		query += fmt.Sprintf(" AND t.card_id = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND card_id = $%d", countArgIdx)
		args = append(args, v)
		countArgs = append(countArgs, v)
		argIdx++
		countArgIdx++
	}
	if v, ok := filters["search"].(string); ok && v != "" {
		query += fmt.Sprintf(" AND t.details ILIKE $%d", argIdx)
		countQuery += fmt.Sprintf(" AND details ILIKE $%d", countArgIdx)
		args = append(args, "%"+v+"%")
		countArgs = append(countArgs, "%"+v+"%")
		argIdx++
		countArgIdx++
	}

	query += " ORDER BY t.executed_at DESC"

	limit := 50
	if lv, ok := filters["limit"].(int); ok && lv > 0 {
		if lv > 500 {
			limit = 500
		} else {
			limit = lv
		}
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)
	argIdx++

	if ov, ok := filters["offset"].(int); ok && ov > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, ov)
	}

	// Total count
	var total int
	if err := GlobalDB.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		log.Printf("Error counting transactions for user %d: %v", userID, err)
		total = 0
	}

	rows, err := GlobalDB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		var cardID sql.NullInt64
		var sourceID sql.NullInt64
		var last4 string
		var executedAt time.Time

		err := rows.Scan(
			&tx.TransactionID, &tx.UserID, &cardID,
			&tx.Amount, &tx.Fee, &tx.TransactionType, &tx.Status,
			&tx.Details, &tx.ProviderTxID,
			&tx.SourceType, &sourceID, &tx.Currency,
			&executedAt,
			&last4,
		)
		if err != nil {
			log.Printf("Error scanning transaction: %v", err)
			continue
		}
		tx.ExecutedAt = executedAt
		if cardID.Valid {
			cid := int(cardID.Int64)
			tx.CardID = &cid
		}
		if sourceID.Valid {
			sid := int(sourceID.Int64)
			tx.SourceID = &sid
		}
		tx.CardLast4Digits = last4
		txs = append(txs, tx)
	}

	if txs == nil {
		txs = []models.Transaction{}
	}
	return txs, total, nil
}
