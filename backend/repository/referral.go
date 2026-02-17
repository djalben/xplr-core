package repository

import (
	"fmt"
	"log"

	"github.com/aalabin/xplr/models"
	"github.com/shopspring/decimal"
)

// GenerateReferralCode - Генерирует уникальный реферальный код
func GenerateReferralCode(userID int) string {
	// Простая генерация: USER{userID}-{random}
	// В продакшене можно использовать более сложную логику
	return fmt.Sprintf("USER%d-%s", userID, generateReferralRandomString(8))
}

// generateReferralRandomString - Генерирует случайную строку для реферального кода
func generateReferralRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// Простая генерация на основе индекса (в продакшене использовать crypto/rand)
		b[i] = charset[(i*7+length)%len(charset)]
	}
	return string(b)
}

// CreateReferral - Создать реферальную запись
func CreateReferral(referrerID int, referredID int, referralCode string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	_, err := GlobalDB.Exec(
		`INSERT INTO referrals (referrer_id, referred_id, referral_code, status)
		 VALUES ($1, $2, $3, 'ACTIVE')`,
		referrerID, referredID, referralCode,
	)
	if err != nil {
		log.Printf("DB Error creating referral: %v", err)
		return fmt.Errorf("failed to create referral")
	}

	log.Printf("✅ Referral created: referrer %d -> referred %d (code: %s)", referrerID, referredID, referralCode)
	return nil
}

// GetUserReferralCode - Получить реферальный код пользователя (или создать новый)
func GetUserReferralCode(userID int) (string, error) {
	if GlobalDB == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	// Проверяем, есть ли уже код
	var code string
	err := GlobalDB.QueryRow(
		"SELECT referral_code FROM referrals WHERE referrer_id = $1 LIMIT 1",
		userID,
	).Scan(&code)

	if err == nil {
		return code, nil
	}

	// Если кода нет, создаем новый
	newCode := GenerateReferralCode(userID)
	
	// Проверяем уникальность
	var existingCode string
	err = GlobalDB.QueryRow(
		"SELECT referral_code FROM referrals WHERE referral_code = $1",
		newCode,
	).Scan(&existingCode)
	
	// Если код уже существует, генерируем новый
	for err == nil {
		newCode = GenerateReferralCode(userID)
		err = GlobalDB.QueryRow(
			"SELECT referral_code FROM referrals WHERE referral_code = $1",
			newCode,
		).Scan(&existingCode)
	}

	return newCode, nil
}

// GetReferralStats - Получить статистику реферальной программы пользователя
func GetReferralStats(userID int) (*models.ReferralStats, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// Получить реферальный код
	code, err := GetUserReferralCode(userID)
	if err != nil {
		return nil, err
	}

	// Подсчитать статистику
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
		TotalCommission:  totalCommission,
		ReferralCode:    code,
	}, nil
}

// ProcessReferralRegistration - Обработать регистрацию по реферальной ссылке
func ProcessReferralRegistration(referredID int, referralCode string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Найти реферера по коду
	var referrerID int
	err := GlobalDB.QueryRow(
		"SELECT referrer_id FROM referrals WHERE referral_code = $1 AND status = 'ACTIVE' LIMIT 1",
		referralCode,
	).Scan(&referrerID)

	if err != nil {
		// Реферальный код не найден или неактивен
		return fmt.Errorf("invalid referral code")
	}

	// Проверяем, что пользователь еще не был приглашен этим реферером
	var existingID int
	err = GlobalDB.QueryRow(
		"SELECT id FROM referrals WHERE referrer_id = $1 AND referred_id = $2",
		referrerID, referredID,
	).Scan(&existingID)

	if err == nil {
		// Уже существует
		return nil // Не ошибка, просто игнорируем
	}

	// Создаем реферальную запись
	return CreateReferral(referrerID, referredID, referralCode)
}

// AddReferralCommission - Добавить комиссию рефереру
func AddReferralCommission(referrerID int, amount decimal.Decimal) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Обновляем commission_earned для всех активных рефералов реферера
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
