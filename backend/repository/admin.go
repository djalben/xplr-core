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
	Role       string `json:"role"`
	IsVerified bool   `json:"is_verified"`
	IsBlocked  bool   `json:"is_blocked"`
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
		       COALESCE(u.role, 'user'), COALESCE(u.is_verified, FALSE),
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
		if err := rows.Scan(&u.ID, &u.Email, &bal, &u.Status, &u.IsAdmin, &u.Role, &u.IsVerified, &u.CardCount, &createdAt); err != nil {
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

// SearchUsersByEmail returns users matching email substring.
func SearchUsersByEmail(query string, limit int) ([]AdminUserRow, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	sql := `
		SELECT u.id, u.email, COALESCE(u.balance_rub, 0), u.status, COALESCE(u.is_admin, FALSE),
		       COALESCE(u.role, 'user'), COALESCE(u.is_verified, FALSE), COALESCE(u.is_blocked, FALSE),
		       (SELECT COUNT(*) FROM cards c WHERE c.user_id = u.id) as card_count,
		       u.created_at
		FROM users u
		WHERE u.email ILIKE '%' || $1 || '%'
		ORDER BY u.id ASC
		LIMIT $2
	`
	rows, err := GlobalDB.Query(sql, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []AdminUserRow
	for rows.Next() {
		var u AdminUserRow
		var bal decimal.Decimal
		var createdAt interface{}
		if err := rows.Scan(&u.ID, &u.Email, &bal, &u.Status, &u.IsAdmin, &u.Role, &u.IsVerified, &u.IsBlocked, &u.CardCount, &createdAt); err != nil {
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

// AdminUpdateUserGrade updates the grade for a given user in user_grades table.
func AdminUpdateUserGrade(userID int, grade string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	valid := map[string]bool{"STANDARD": true, "SILVER": true, "GOLD": true, "PLATINUM": true, "BLACK": true}
	if !valid[grade] {
		return fmt.Errorf("invalid grade: %s", grade)
	}
	_, err := GlobalDB.Exec(
		`UPDATE user_grades SET grade = $1, updated_at = NOW() WHERE user_id = $2`,
		grade, userID,
	)
	if err != nil {
		log.Printf("AdminUpdateUserGrade: error: %v", err)
		return err
	}
	log.Printf("✅ Admin set user %d grade to %s", userID, grade)
	return nil
}

// WriteAdminLog records an admin action.
func WriteAdminLog(adminID int, action string) {
	if GlobalDB == nil {
		return
	}
	_, err := GlobalDB.Exec(
		`INSERT INTO admin_logs (admin_id, action) VALUES ($1, $2)`,
		adminID, action,
	)
	if err != nil {
		log.Printf("WriteAdminLog: error: %v", err)
	}
}

// GetAdminLogs returns recent admin log entries.
func GetAdminLogs(limit int) ([]map[string]interface{}, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := GlobalDB.Query(
		`SELECT al.id, al.admin_id, COALESCE(u.email, ''), al.action, al.created_at
		 FROM admin_logs al LEFT JOIN users u ON u.id = al.admin_id
		 ORDER BY al.created_at DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []map[string]interface{}
	for rows.Next() {
		var id, adminID int
		var email, action string
		var createdAt time.Time
		if err := rows.Scan(&id, &adminID, &email, &action, &createdAt); err != nil {
			continue
		}
		logs = append(logs, map[string]interface{}{
			"id":          id,
			"admin_id":    adminID,
			"admin_email": email,
			"action":      action,
			"created_at":  createdAt,
		})
	}
	if logs == nil {
		logs = []map[string]interface{}{}
	}
	return logs, nil
}

// --- Commission Config ---

type CommissionConfigRow struct {
	ID          int    `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at"`
}

func GetAllCommissionConfigs() ([]CommissionConfigRow, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	rows, err := GlobalDB.Query(`SELECT id, key, value, COALESCE(description, ''), updated_at FROM commission_config ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var configs []CommissionConfigRow
	for rows.Next() {
		var c CommissionConfigRow
		var val decimal.Decimal
		var updatedAt time.Time
		if err := rows.Scan(&c.ID, &c.Key, &val, &c.Description, &updatedAt); err != nil {
			continue
		}
		c.Value = val.String()
		c.UpdatedAt = updatedAt.Format(time.RFC3339)
		configs = append(configs, c)
	}
	if configs == nil {
		configs = []CommissionConfigRow{}
	}
	return configs, nil
}

func UpdateCommissionConfig(id int, value decimal.Decimal) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE commission_config SET value = $1, updated_at = NOW() WHERE id = $2`,
		value, id,
	)
	return err
}

func GetCommissionValue(key string) (decimal.Decimal, error) {
	if GlobalDB == nil {
		return decimal.Zero, fmt.Errorf("database connection not initialized")
	}
	var val decimal.Decimal
	err := GlobalDB.QueryRow(`SELECT value FROM commission_config WHERE key = $1`, key).Scan(&val)
	return val, err
}

// --- Support Tickets ---

type SupportTicketRow struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	TgChatID  *int64 `json:"tg_chat_id"`
	CreatedAt string `json:"created_at"`
}

func GetAllSupportTickets(statusFilter string) ([]SupportTicketRow, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	q := `SELECT st.id, st.user_id, COALESCE(st.email, u.email), st.subject, COALESCE(st.message, ''), st.status, st.tg_chat_id, st.created_at
	      FROM support_tickets st JOIN users u ON u.id = st.user_id`
	var args []interface{}
	if statusFilter != "" {
		q += ` WHERE st.status = $1`
		args = append(args, statusFilter)
	}
	q += ` ORDER BY st.created_at DESC LIMIT 100`
	rows, err := GlobalDB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tickets []SupportTicketRow
	for rows.Next() {
		var t SupportTicketRow
		var createdAt time.Time
		if err := rows.Scan(&t.ID, &t.UserID, &t.Email, &t.Subject, &t.Message, &t.Status, &t.TgChatID, &createdAt); err != nil {
			continue
		}
		t.CreatedAt = createdAt.Format(time.RFC3339)
		tickets = append(tickets, t)
	}
	if tickets == nil {
		tickets = []SupportTicketRow{}
	}
	return tickets, nil
}

// CreateSupportTicket inserts a new support ticket and returns its ID.
func CreateSupportTicket(userID int, email, subject, message string) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}
	var id int
	err := GlobalDB.QueryRow(
		`INSERT INTO support_tickets (user_id, email, subject, message, status) VALUES ($1, $2, $3, $4, 'open') RETURNING id`,
		userID, email, subject, message,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	log.Printf("[SUPPORT] Ticket #%d created by user %d (%s)", id, userID, email)
	return id, nil
}

func UpdateSupportTicketStatus(ticketID int, status string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE support_tickets SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, ticketID,
	)
	return err
}

// --- Emergency Freeze ---

// EmergencyFreezeUser freezes all cards and zeroes wallet for a user.
func EmergencyFreezeUser(userID int) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}
	// 1. Freeze all active cards
	res, err := GlobalDB.Exec(
		`UPDATE cards SET card_status = 'FROZEN' WHERE user_id = $1 AND card_status = 'ACTIVE'`,
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to freeze cards: %w", err)
	}
	frozenCount, _ := res.RowsAffected()

	// 2. Ban user
	_, err = GlobalDB.Exec(`UPDATE users SET status = 'BANNED' WHERE id = $1`, userID)
	if err != nil {
		return int(frozenCount), fmt.Errorf("cards frozen but failed to ban user: %w", err)
	}

	// 3. Zero out wallet balance
	_, _ = GlobalDB.Exec(`UPDATE internal_balances SET master_balance = 0, updated_at = NOW() WHERE user_id = $1`, userID)
	_, _ = GlobalDB.Exec(`UPDATE users SET balance_rub = 0 WHERE id = $1`, userID)

	log.Printf("🚨 EMERGENCY FREEZE: user %d — %d cards frozen, status=BANNED, balance zeroed", userID, frozenCount)
	return int(frozenCount), nil
}

// --- Enhanced Admin Stats ---

type AdminDashboardStats struct {
	TotalUsers   int    `json:"total_users"`
	TotalBalance string `json:"total_balance"`
	ActiveCards  int    `json:"active_cards"`
	OpenTickets  int    `json:"open_tickets"`
	TodaySignups int    `json:"today_signups"`
	TotalCards   int    `json:"total_cards"`
}

func GetAdminDashboardStats() (*AdminDashboardStats, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	s := &AdminDashboardStats{}
	GlobalDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&s.TotalUsers)
	var bal decimal.Decimal
	GlobalDB.QueryRow("SELECT COALESCE(SUM(balance_rub), 0) FROM users").Scan(&bal)
	s.TotalBalance = bal.String()
	GlobalDB.QueryRow("SELECT COUNT(*) FROM cards WHERE card_status='ACTIVE'").Scan(&s.ActiveCards)
	GlobalDB.QueryRow("SELECT COUNT(*) FROM cards").Scan(&s.TotalCards)
	GlobalDB.QueryRow("SELECT COUNT(*) FROM support_tickets WHERE status IN ('open','in_progress')").Scan(&s.OpenTickets)
	GlobalDB.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE").Scan(&s.TodaySignups)
	return s, nil
}

// ToggleUserBlock flips is_blocked for a user and returns the new value.
func ToggleUserBlock(userID int) (bool, string, error) {
	if GlobalDB == nil {
		return false, "", fmt.Errorf("database connection not initialized")
	}
	var newBlocked bool
	var email string
	err := GlobalDB.QueryRow(
		`UPDATE users SET is_blocked = NOT COALESCE(is_blocked, FALSE) WHERE id = $1
		 RETURNING is_blocked, email`, userID,
	).Scan(&newBlocked, &email)
	if err != nil {
		return false, "", fmt.Errorf("user not found or update failed: %v", err)
	}
	return newBlocked, email, nil
}

// IsUserBlocked checks if a user is blocked. Returns false on any error.
func IsUserBlocked(userID int) bool {
	if GlobalDB == nil {
		return false
	}
	var blocked bool
	err := GlobalDB.QueryRow(
		`SELECT COALESCE(is_blocked, FALSE) FROM users WHERE id = $1`, userID,
	).Scan(&blocked)
	if err != nil {
		return false
	}
	return blocked
}
