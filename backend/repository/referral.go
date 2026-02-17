package repository

import (
	cryptorand "crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/notification"
	"github.com/shopspring/decimal"
)

// generateSecureCode generates a truly random alphanumeric code of given length.
func generateSecureCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, length)
	for i := range b {
		n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			b[i] = charset[i%len(charset)]
			continue
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// GetUserReferralCode returns the user's persistent referral code, creating one if needed.
func GetUserReferralCode(userID int) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	// Check referral_codes table first
	var code string
	err := GlobalDB.QueryRow(
		"SELECT code FROM referral_codes WHERE user_id = $1",
		userID,
	).Scan(&code)
	if err == nil {
		return code, nil
	}

	// Fallback: check legacy referrals table
	err = GlobalDB.QueryRow(
		"SELECT referral_code FROM referrals WHERE referrer_id = $1 LIMIT 1",
		userID,
	).Scan(&code)
	if err == nil {
		// Persist to referral_codes
		GlobalDB.Exec("INSERT INTO referral_codes (user_id, code) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, code)
		return code, nil
	}

	// Generate new code and persist
	for attempts := 0; attempts < 10; attempts++ {
		newCode := generateSecureCode(8)
		_, err = GlobalDB.Exec(
			"INSERT INTO referral_codes (user_id, code) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			userID, newCode,
		)
		if err == nil {
			log.Printf("✅ Referral code %s created for user %d", newCode, userID)
			return newCode, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique referral code")
}

// GetReferralStats returns referral program statistics for a user.
func GetReferralStats(userID int) (*models.ReferralStats, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	code, err := GetUserReferralCode(userID)
	if err != nil {
		return nil, err
	}

	var totalReferrals, activeReferrals int
	var totalCommission decimal.Decimal

	err = GlobalDB.QueryRow(
		`SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'ACTIVE' THEN 1 END) as active,
			COALESCE(SUM(commission_earned), 0) as commission
		 FROM referrals 
		 WHERE referrer_id = $1`,
		userID,
	).Scan(&totalReferrals, &activeReferrals, &totalCommission)

	if err != nil {
		log.Printf("DB Error fetching referral stats: %v", err)
		return nil, fmt.Errorf("failed to fetch referral stats")
	}

	return &models.ReferralStats{
		TotalReferrals:  totalReferrals,
		ActiveReferrals: activeReferrals,
		TotalCommission: totalCommission,
		ReferralCode:    code,
	}, nil
}

// ProcessReferralRegistration handles registration via a referral link.
// It looks up the referrer from referral_codes, creates the referral record,
// credits $5 bonus to the new user, and sends Telegram notification to the referrer.
func ProcessReferralRegistration(referredID int, referralCode string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	referralCode = strings.TrimSpace(strings.ToUpper(referralCode))
	if referralCode == "" {
		return fmt.Errorf("empty referral code")
	}

	// Look up referrer from referral_codes table
	var referrerID int
	err := GlobalDB.QueryRow(
		"SELECT user_id FROM referral_codes WHERE code = $1",
		referralCode,
	).Scan(&referrerID)

	if err != nil {
		// Fallback: try legacy referrals table
		err = GlobalDB.QueryRow(
			"SELECT referrer_id FROM referrals WHERE referral_code = $1 LIMIT 1",
			referralCode,
		).Scan(&referrerID)
		if err != nil {
			log.Printf("Referral code %s not found in any table", referralCode)
			return fmt.Errorf("invalid referral code")
		}
	}

	// Cannot refer yourself
	if referrerID == referredID {
		return fmt.Errorf("cannot refer yourself")
	}

	// Check for duplicate
	var existingID int
	err = GlobalDB.QueryRow(
		"SELECT id FROM referrals WHERE referrer_id = $1 AND referred_id = $2",
		referrerID, referredID,
	).Scan(&existingID)
	if err == nil {
		return nil // already exists, silently ignore
	}

	// Create referral record
	_, err = GlobalDB.Exec(
		`INSERT INTO referrals (referrer_id, referred_id, referral_code, status)
		 VALUES ($1, $2, $3, 'ACTIVE')`,
		referrerID, referredID, referralCode,
	)
	if err != nil {
		log.Printf("DB Error creating referral: %v", err)
		return fmt.Errorf("failed to create referral")
	}
	log.Printf("✅ Referral created: referrer %d -> referred %d (code: %s)", referrerID, referredID, referralCode)

	// Credit $5 bonus to the new user
	bonus := decimal.NewFromInt(5)
	_, err = GlobalDB.Exec(
		"UPDATE users SET balance_rub = COALESCE(balance_rub, 0) + $1, balance = COALESCE(balance, 0) + $1 WHERE id = $2",
		bonus, referredID,
	)
	if err != nil {
		log.Printf("Warning: failed to credit referral bonus to user %d: %v", referredID, err)
	} else {
		log.Printf("✅ $5 referral bonus credited to new user %d", referredID)
	}

	// Send Telegram notification to referrer
	go func() {
		referrer, err := GetUserByID(referrerID)
		if err != nil {
			return
		}
		referred, err := GetUserByID(referredID)
		if err != nil {
			return
		}
		if referrer.TelegramChatID.Valid {
			msg := fmt.Sprintf("🎉 Новый реферал!\n\nПользователь %s зарегистрировался по вашей ссылке.\nВаш код: %s",
				referred.Email, referralCode)
			notification.SendTelegramMessage(referrer.TelegramChatID.Int64, msg)
		}
	}()

	return nil
}

// GetReferrerID returns the referrer_id for a user, or 0 if none.
func GetReferrerID(userID int) int {
	if GlobalDB == nil {
		return 0
	}
	var referrerID int
	err := GlobalDB.QueryRow(
		"SELECT referrer_id FROM referrals WHERE referred_id = $1 AND status = 'ACTIVE' LIMIT 1",
		userID,
	).Scan(&referrerID)
	if err != nil {
		return 0
	}
	return referrerID
}

// CreditRevShare atomically credits RevShare commission to the referrer and records it.
// commission = amount * 5%.  All in one DB transaction.
func CreditRevShare(referrerID int, sourceUserID int, sourceAmount decimal.Decimal, description string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	commission := sourceAmount.Mul(decimal.NewFromFloat(0.05)).Round(2)
	if commission.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	tx, err := GlobalDB.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// 1. Credit referrer balance atomically
	_, err = tx.Exec(
		"UPDATE users SET balance_rub = COALESCE(balance_rub, 0) + $1, balance = COALESCE(balance, 0) + $1 WHERE id = $2",
		commission, referrerID,
	)
	if err != nil {
		return fmt.Errorf("credit balance: %v", err)
	}

	// 2. Record REFERRAL_REVENUE transaction
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, amount, fee, transaction_type, status, details, executed_at)
		 VALUES ($1, $2, 0, 'REFERRAL_REVENUE', 'APPROVED', $3, NOW())`,
		referrerID, commission,
		fmt.Sprintf("RevShare 5%%: %s (from user %d)", description, sourceUserID),
	)
	if err != nil {
		return fmt.Errorf("record tx: %v", err)
	}

	// 3. Update commission_earned in referrals table
	_, err = tx.Exec(
		`UPDATE referrals SET commission_earned = COALESCE(commission_earned, 0) + $1 WHERE referrer_id = $2 AND referred_id = $3`,
		commission, referrerID, sourceUserID,
	)
	if err != nil {
		log.Printf("Warning: failed to update referrals commission_earned: %v", err)
		// non-fatal
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %v", err)
	}

	log.Printf("✅ RevShare $%s credited to referrer %d (from user %d: %s)", commission.String(), referrerID, sourceUserID, description)
	return nil
}

// AddReferralCommission adds commission to referrer's referral records.
func AddReferralCommission(referrerID int, amount decimal.Decimal) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	_, err := GlobalDB.Exec(
		`UPDATE referrals 
		 SET commission_earned = commission_earned + $1 
		 WHERE referrer_id = $2 AND status = 'ACTIVE'`,
		amount, referrerID,
	)
	if err != nil {
		log.Printf("DB Error adding referral commission: %v", err)
		return fmt.Errorf("failed to add referral commission")
	}

	return nil
}

// GetReferralList returns the list of referred users for a referrer.
func GetReferralList(userID int) ([]map[string]interface{}, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	rows, err := GlobalDB.Query(
		`SELECT r.referred_id, u.email, r.status, COALESCE(r.commission_earned, 0), r.created_at
		 FROM referrals r
		 JOIN users u ON u.id = r.referred_id
		 WHERE r.referrer_id = $1
		 ORDER BY r.created_at DESC`,
		userID,
	)
	if err != nil {
		log.Printf("DB Error fetching referral list: %v", err)
		return nil, err
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var refID int
		var email, status string
		var commission decimal.Decimal
		var createdAt sql.NullTime
		if err := rows.Scan(&refID, &email, &status, &commission, &createdAt); err != nil {
			continue
		}
		ca := ""
		if createdAt.Valid {
			ca = createdAt.Time.Format("2006-01-02")
		}
		list = append(list, map[string]interface{}{
			"id":         refID,
			"email":      email,
			"status":     status,
			"commission": commission.String(),
			"created_at": ca,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, nil
}
