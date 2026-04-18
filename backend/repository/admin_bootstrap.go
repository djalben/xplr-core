package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/djalben/xplr-core/backend/domain"
)

// HasAnyAdmin returns true if at least one user with is_admin=true or role='admin' exists.
func HasAnyAdmin() bool {
	if GlobalDB == nil {
		return false
	}
	var count int
	err := GlobalDB.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin = TRUE OR role = 'admin'`).Scan(&count)
	if err != nil {
		log.Printf("[BOOTSTRAP] Warning: could not check for admins: %v", err)
		return true // Assume there's an admin to avoid accidental promotion
	}
	return count > 0
}

// PromoteToAdmin sets is_admin=true and role='admin' for the given user.
func PromoteToAdmin(userID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE users SET is_admin = TRUE, role = 'admin' WHERE id = $1`, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to promote user %d: %w", userID, err)
	}
	log.Printf("[BOOTSTRAP] User %d promoted to admin", userID)
	return nil
}

// GetUserByIDBasic fetches only essential columns for a user by ID.
// Fallback when the full GetUserByID fails due to schema issues.
func GetUserByIDBasic(userID int) (domain.User, error) {
	if GlobalDB == nil {
		return domain.User{}, fmt.Errorf("database connection not initialized")
	}

	query := `SELECT id, email, password_hash, COALESCE(status, 'ACTIVE'), COALESCE(is_admin, FALSE), COALESCE(role, 'user') FROM users WHERE id = $1`
	var user domain.User
	err := GlobalDB.QueryRow(query, userID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Status, &user.IsAdmin, &user.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.New("пользователь не найден")
		}
		return domain.User{}, err
	}
	return user, nil
}
