package repository

import (
	"fmt"
	"log"

	"github.com/aalabin/xplr/models"
	"github.com/shopspring/decimal"
)

// Grade thresholds (как у e.pn)
var (
	GradeStandard = "STANDARD" // При регистрации
	GradeSilver   = "SILVER"   // $1,000
	GradeGold     = "GOLD"     // $10,000
	GradePlatinum = "PLATINUM" // $50,000
	GradeBlack    = "BLACK"    // $100,000
)

// Grade fee percentages (как у e.pn)
var (
	FeeStandard = decimal.NewFromFloat(6.70) // 6.7%
	FeeSilver   = decimal.NewFromFloat(6.00) // 6.0%
	FeeGold     = decimal.NewFromFloat(5.00) // 5.0%
	FeePlatinum = decimal.NewFromFloat(4.00) // 4.0%
	FeeBlack    = decimal.NewFromFloat(3.00) // 3.0%
)

// GetUserGrade - Получить Grade пользователя
func GetUserGrade(userID int) (*models.UserGrade, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var grade models.UserGrade
	err := GlobalDB.QueryRow(
		"SELECT id, user_id, grade, total_spent, fee_percent, updated_at FROM user_grades WHERE user_id = $1",
		userID,
	).Scan(&grade.ID, &grade.UserID, &grade.Grade, &grade.TotalSpent, &grade.FeePercent, &grade.UpdatedAt)

	if err != nil {
		// Если Grade не найден, создаем STANDARD
		return CreateUserGrade(userID)
	}

	return &grade, nil
}

// CreateUserGrade - Создать Grade для пользователя (STANDARD по умолчанию)
func CreateUserGrade(userID int) (*models.UserGrade, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var grade models.UserGrade
	err := GlobalDB.QueryRow(
		`INSERT INTO user_grades (user_id, grade, total_spent, fee_percent) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, user_id, grade, total_spent, fee_percent, updated_at`,
		userID, GradeStandard, decimal.Zero, FeeStandard,
	).Scan(&grade.ID, &grade.UserID, &grade.Grade, &grade.TotalSpent, &grade.FeePercent, &grade.UpdatedAt)

	if err != nil {
		log.Printf("DB Error creating user grade: %v", err)
		return nil, fmt.Errorf("failed to create user grade")
	}

	return &grade, nil
}

// CalculateGradeFromSpent - Вычислить Grade на основе суммы трат
func CalculateGradeFromSpent(totalSpent decimal.Decimal) (string, decimal.Decimal) {
	// Пороги как у e.pn
	thresholdBlack := decimal.NewFromInt(100000)    // $100,000
	thresholdPlatinum := decimal.NewFromInt(50000)  // $50,000
	thresholdGold := decimal.NewFromInt(10000)      // $10,000
	thresholdSilver := decimal.NewFromInt(1000)      // $1,000

	if totalSpent.GreaterThanOrEqual(thresholdBlack) {
		return GradeBlack, FeeBlack
	} else if totalSpent.GreaterThanOrEqual(thresholdPlatinum) {
		return GradePlatinum, FeePlatinum
	} else if totalSpent.GreaterThanOrEqual(thresholdGold) {
		return GradeGold, FeeGold
	} else if totalSpent.GreaterThanOrEqual(thresholdSilver) {
		return GradeSilver, FeeSilver
	}
	return GradeStandard, FeeStandard
}

// UpdateUserGrade - Обновить Grade пользователя на основе трат
func UpdateUserGrade(userID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Вычислить общую сумму трат пользователя (только APPROVED транзакции типа CAPTURE)
	var totalSpent decimal.Decimal
	err := GlobalDB.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) 
		 FROM transactions 
		 WHERE user_id = $1 
		   AND transaction_type = 'CAPTURE' 
		   AND status = 'APPROVED'`,
		userID,
	).Scan(&totalSpent)

	if err != nil {
		log.Printf("DB Error calculating total spent for user %d: %v", userID, err)
		return fmt.Errorf("failed to calculate total spent")
	}

	// Вычислить новый Grade
	newGrade, newFeePercent := CalculateGradeFromSpent(totalSpent)

	// Обновить или создать Grade
	_, err = GlobalDB.Exec(
		`INSERT INTO user_grades (user_id, grade, total_spent, fee_percent, updated_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (user_id) 
		 DO UPDATE SET 
		   grade = $2,
		   total_spent = $3,
		   fee_percent = $4,
		   updated_at = NOW()`,
		userID, newGrade, totalSpent, newFeePercent,
	)

	if err != nil {
		log.Printf("DB Error updating user grade: %v", err)
		return fmt.Errorf("failed to update user grade")
	}

	log.Printf("✅ User %d grade updated: %s (total spent: %s, fee: %s%%)", 
		userID, newGrade, totalSpent.String(), newFeePercent.String())
	return nil
}

// GetUserGradeInfo - Получить информацию о Grade с деталями для UI
func GetUserGradeInfo(userID int) (*models.GradeInfo, error) {
	grade, err := GetUserGrade(userID)
	if err != nil {
		return nil, err
	}

	info := &models.GradeInfo{
		Grade:      grade.Grade,
		TotalSpent: grade.TotalSpent,
		FeePercent: grade.FeePercent,
	}

	// Вычислить следующий уровень
	thresholdBlack := decimal.NewFromInt(100000)
	thresholdPlatinum := decimal.NewFromInt(50000)
	thresholdGold := decimal.NewFromInt(10000)
	thresholdSilver := decimal.NewFromInt(1000)

	var nextGrade string
	var nextSpend decimal.Decimal

	switch grade.Grade {
	case GradeStandard:
		nextGrade = GradeSilver
		nextSpend = thresholdSilver.Sub(grade.TotalSpent)
	case GradeSilver:
		nextGrade = GradeGold
		nextSpend = thresholdGold.Sub(grade.TotalSpent)
	case GradeGold:
		nextGrade = GradePlatinum
		nextSpend = thresholdPlatinum.Sub(grade.TotalSpent)
	case GradePlatinum:
		nextGrade = GradeBlack
		nextSpend = thresholdBlack.Sub(grade.TotalSpent)
	case GradeBlack:
		// Максимальный уровень, следующего нет
		nextGrade = ""
		nextSpend = decimal.Zero
	}

	if nextGrade != "" {
		info.NextGrade = &nextGrade
		info.NextSpend = &nextSpend
	}

	return info, nil
}
