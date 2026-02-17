package repository

import (
	"fmt"
	"log"

	"github.com/shopspring/decimal"
)

// AdminStats holds aggregate platform statistics.
type AdminStats struct {
	TotalUsers    int    `json:"total_users"`
	TotalBalance  string `json:"total_balance"`
	TotalCards    int    `json:"total_cards"`
	ActiveCards   int    `json:"active_cards"`
	FrozenCards   int    `json:"frozen_cards"`
	ClosedCards   int    `json:"closed_cards"`
	BlockedCards  int    `json:"blocked_cards"`
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
