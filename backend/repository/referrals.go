package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// SyncUserReferralCode — записывает реферальный код в поле users.referral_code (если ещё пусто).
func SyncUserReferralCode(userID int, code string) {
	if GlobalDB == nil {
		return
	}
	_, _ = GlobalDB.Exec(
		`UPDATE users SET referral_code = $1 WHERE id = $2 AND (referral_code IS NULL OR referral_code = '')`,
		code, userID,
	)
}

// SetReferredBy — устанавливает кто пригласил пользователя (users.referred_by).
func SetReferredBy(userID, referrerID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE users SET referred_by = $1 WHERE id = $2 AND referred_by IS NULL`,
		referrerID, userID,
	)
	return err
}

// RecentReferral — запись о последнем реферале для API.
type RecentReferral struct {
	ID         int             `json:"id"`
	Email      string          `json:"email"`
	JoinedDate time.Time       `json:"joined_date"`
	IsActive   bool            `json:"is_active"`
	Earnings   decimal.Decimal `json:"earnings"`
}

// GetRecentReferralsV2 — топ-5 последних рефералов из legacy referrals table.
func GetRecentReferralsV2(userID int) ([]RecentReferral, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	rows, err := GlobalDB.Query(`
		SELECT r.referred_id, u.email, r.created_at, r.status, COALESCE(r.commission_earned, 0)
		FROM referrals r
		JOIN users u ON u.id = r.referred_id
		WHERE r.referrer_id = $1
		ORDER BY r.created_at DESC
		LIMIT 5
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent referrals: %w", err)
	}
	defer rows.Close()

	var refs []RecentReferral
	for rows.Next() {
		var r RecentReferral
		var status string
		var earningsStr string
		var createdAt sql.NullTime
		if err := rows.Scan(&r.ID, &r.Email, &createdAt, &status, &earningsStr); err != nil {
			log.Printf("Error scanning referral: %v", err)
			continue
		}
		if createdAt.Valid {
			r.JoinedDate = createdAt.Time
		}
		r.IsActive = status == "ACTIVE"
		r.Earnings, _ = decimal.NewFromString(earningsStr)
		// Маскируем email для приватности
		parts := strings.Split(r.Email, "@")
		if len(parts) == 2 && len(parts[0]) > 2 {
			r.Email = parts[0][:2] + "***@" + parts[1]
		}
		refs = append(refs, r)
	}
	if refs == nil {
		refs = []RecentReferral{}
	}
	return refs, nil
}

// ProcessReferralCommissionV2 — начисляет комиссию через referral_commissions + wallet.
func ProcessReferralCommissionV2(referredUserID int, topupAmount decimal.Decimal, sourceTxID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	referrerID := GetReferrerID(referredUserID)
	if referrerID == 0 {
		return nil
	}

	commissionPercent := decimal.NewFromFloat(5.0)
	commission := topupAmount.Mul(commissionPercent).Div(decimal.NewFromInt(100))
	if commission.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	// Записываем комиссию в referral_commissions
	_, err := GlobalDB.Exec(`
		INSERT INTO referral_commissions (referrer_id, referred_id, source_transaction_id, commission_amount, commission_percent)
		VALUES ($1, $2, $3, $4, $5)
	`, referrerID, referredUserID, sourceTxID, commission, commissionPercent)
	if err != nil {
		log.Printf("[REFERRAL] Warning: failed to record commission: %v", err)
	}

	// Начисляем через CreditRevShare (уже обновляет баланс + транзакцию)
	return CreditRevShare(referrerID, referredUserID, topupAmount, "пополнение кошелька")
}
