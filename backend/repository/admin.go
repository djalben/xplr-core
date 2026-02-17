package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// AdminStats holds aggregate platform statistics.
type AdminStats struct {
	TotalUsers   int    `json:"total_users"`
	TotalBalance string `json:"total_balance"`
	TotalCards   int    `json:"total_cards"`
	ActiveCards  int    `json:"active_cards"`
	FrozenCards  int    `json:"frozen_cards"`
	ClosedCards  int    `json:"closed_cards"`
	BlockedCards int    `json:"blocked_cards"`
}

// AdminUserRow is a simplified user for admin listing.
type AdminUserRow struct {
	ID         int    `json:"id"`
	Email      string `json:"email"`
	BalanceRub string `json:"balance_rub"`
	Status     string `json:"status"`
	IsAdmin    bool   `json:"is_admin"`
	CardCount  int    `json:"card_count"`
	CreatedAt  string `json:"created_at"`
}

// GetAdminStats returns aggregate platform stats.
func GetAdminStats() (*AdminStats, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	stats := &AdminStats{}

	// Total users
	err := GlobalDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		log.Printf("AdminStats: error counting users: %v", err)
		return nil, err
	}

	// Total balance
	var totalBal decimal.Decimal
	err = GlobalDB.QueryRow("SELECT COALESCE(SUM(balance_rub), 0) FROM users").Scan(&totalBal)
	if err != nil {
		log.Printf("AdminStats: error summing balances: %v", err)
		return nil, err
	}
	stats.TotalBalance = totalBal.String()

	// Card counts
	err = GlobalDB.QueryRow("SELECT COUNT(*) FROM cards").Scan(&stats.TotalCards)
	if err != nil {
		log.Printf("AdminStats: error counting cards: %v", err)
		return nil, err
	}

	GlobalDB.QueryRow("SELECT COUNT(*) FROM cards WHERE card_status='ACTIVE'").Scan(&stats.ActiveCards)
	GlobalDB.QueryRow("SELECT COUNT(*) FROM cards WHERE card_status='FROZEN'").Scan(&stats.FrozenCards)
	GlobalDB.QueryRow("SELECT COUNT(*) FROM cards WHERE card_status='CLOSED'").Scan(&stats.ClosedCards)
	GlobalDB.QueryRow("SELECT COUNT(*) FROM cards WHERE card_status='BLOCKED'").Scan(&stats.BlockedCards)

	return stats, nil
}

// GetAllUsersForAdmin returns all users with card counts.
func GetAllUsersForAdmin() ([]AdminUserRow, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT u.id, u.email, COALESCE(u.balance_rub, 0), u.status, COALESCE(u.is_admin, FALSE),
		       (SELECT COUNT(*) FROM cards c WHERE c.user_id = u.id) as card_count,
		       u.created_at
		FROM users u
		ORDER BY u.id ASC
	`
	rows, err := GlobalDB.Query(query)
	if err != nil {
		log.Printf("AdminUsers: error querying users: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []AdminUserRow
	for rows.Next() {
		var u AdminUserRow
		var bal decimal.Decimal
		var createdAt interface{}
		if err := rows.Scan(&u.ID, &u.Email, &bal, &u.Status, &u.IsAdmin, &u.CardCount, &createdAt); err != nil {
			log.Printf("AdminUsers: error scanning row: %v", err)
			continue
		}
		u.BalanceRub = bal.String()
		u.CreatedAt = fmt.Sprintf("%v", createdAt)
		users = append(users, u)
	}
	if users == nil {
		users = []AdminUserRow{}
	}
	return users, nil
}

// AdminAdjustBalance adds (or subtracts if negative) amount to a user's balance.
func AdminAdjustBalance(targetUserID int, amount decimal.Decimal) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf("database connection not initialized")
	}
	tx, err := GlobalDB.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction")
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"UPDATE users SET balance_rub = COALESCE(balance_rub, 0) + $1, balance = COALESCE(balance, 0) + $1 WHERE id = $2",
		amount, targetUserID,
	)
	if err != nil {
		log.Printf("AdminAdjustBalance: DB error: %v", err)
		return "", fmt.Errorf("failed to adjust balance")
	}

	details := fmt.Sprintf("Admin balance adjustment: %s", amount.String())
	txType := "FUND"
	if amount.LessThan(decimal.Zero) {
		txType = "CAPTURE"
	}
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, amount, fee, transaction_type, status, details, executed_at)
		 VALUES ($1, $2, 0, $3, 'APPROVED', $4, $5)`,
		targetUserID, amount.Abs(), txType, details, time.Now(),
	)
	if err != nil {
		log.Printf("AdminAdjustBalance: failed to log transaction: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit")
	}

	// Fetch new balance
	var newBal decimal.Decimal
	GlobalDB.QueryRow("SELECT COALESCE(balance_rub, 0) FROM users WHERE id = $1", targetUserID).Scan(&newBal)
	log.Printf("✅ Admin adjusted user %d balance by %s. New balance: %s", targetUserID, amount.String(), newBal.String())
	return newBal.String(), nil
}

// AdminToggleRole toggles is_admin for a user.
func AdminToggleRole(targetUserID int) (bool, error) {
	if GlobalDB == nil {
		return false, fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		"UPDATE users SET is_admin = NOT COALESCE(is_admin, FALSE) WHERE id = $1",
		targetUserID,
	)
	if err != nil {
		log.Printf("AdminToggleRole: DB error: %v", err)
		return false, fmt.Errorf("failed to toggle role")
	}
	var newVal bool
	GlobalDB.QueryRow("SELECT COALESCE(is_admin, FALSE) FROM users WHERE id = $1", targetUserID).Scan(&newVal)
	log.Printf("✅ Admin toggled user %d is_admin to %v", targetUserID, newVal)
	return newVal, nil
}

// AdminSetUserStatus sets user status to ACTIVE or BANNED.
func AdminSetUserStatus(targetUserID int, status string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	if status != "ACTIVE" && status != "BANNED" {
		return fmt.Errorf("invalid status: must be ACTIVE or BANNED")
	}
	_, err := GlobalDB.Exec(
		"UPDATE users SET status = $1 WHERE id = $2",
		status, targetUserID,
	)
	if err != nil {
		log.Printf("AdminSetUserStatus: DB error: %v", err)
		return fmt.Errorf("failed to update user status")
	}
	log.Printf("✅ Admin set user %d status to %s", targetUserID, status)
	return nil
}
