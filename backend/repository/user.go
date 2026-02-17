package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/notification"
	"github.com/shopspring/decimal"
)

// GlobalDB должен быть объявлен в этом пакете (например, globals.go)

// --- ФУНКЦИИ АУТЕНТИФИКАЦИИ И ПОЛУЧЕНИЯ ДАННЫХ ---

// CreateUser создает нового пользователя и сохраняет его в БД.
func CreateUser(user models.User) (models.User, error) {
	if GlobalDB == nil {
		return models.User{}, fmt.Errorf("database connection not initialized")
	}

	queryUser := `
		INSERT INTO users (email, password_hash, balance, balance_rub, kyc_status, active_mode, status) 
		VALUES ($1, $2, 0.00, 0.00, 'pending', 'personal', 'ACTIVE') 
		RETURNING id, created_at, balance, balance_rub
	`
	var createdUser models.User

	err := GlobalDB.QueryRow(queryUser, user.Email, user.PasswordHash).
		Scan(&createdUser.ID, &createdUser.CreatedAt, &createdUser.Balance, &createdUser.BalanceRub)

	if err != nil {
		log.Printf("Error creating user %s: %v", user.Email, err)
		return models.User{}, err
	}

	createdUser.Email = user.Email

	log.Printf("User %s created with ID: %d", createdUser.Email, createdUser.ID)

	// Генерируем API-ключ для нового пользователя.
	_, err = GenerateAPIKey(createdUser.ID)
	if err != nil {
		log.Printf("WARNING: Could not generate API key for user %d on creation: %v", createdUser.ID, err)
	}

	return createdUser, nil
}

// GetUserByEmail - Находит пользователя по email.
func GetUserByEmail(email string) (models.User, error) {
	if GlobalDB == nil {
		return models.User{}, fmt.Errorf("database connection not initialized")
	}

	query := `SELECT id, email, password_hash, balance, COALESCE(balance_rub, 0), COALESCE(balance_arbitrage, 0), COALESCE(balance_personal, 0), COALESCE(kyc_status, ''), COALESCE(active_mode, 'personal'), created_at, status, telegram_chat_id, COALESCE(is_admin, FALSE) FROM users WHERE email = $1`

	var user models.User

	err := GlobalDB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Balance,
		&user.BalanceRub,
		&user.BalanceArbitrage,
		&user.BalancePersonal,
		&user.KYCStatus,
		&user.ActiveMode,
		&user.CreatedAt,
		&user.Status,
		&user.TelegramChatID,
		&user.IsAdmin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("пользователь не найден")
		}
		log.Printf("DB Error GetUserByEmail: %v", err)
		return models.User{}, err
	}

	return user, nil
}

// GetUserByID - Находит пользователя по ID.
func GetUserByID(userID int) (models.User, error) {
	if GlobalDB == nil {
		return models.User{}, fmt.Errorf("database connection not initialized")
	}

	query := `SELECT id, email, password_hash, balance, COALESCE(balance_rub, 0), COALESCE(balance_arbitrage, 0), COALESCE(balance_personal, 0), COALESCE(kyc_status, ''), COALESCE(active_mode, 'personal'), created_at, status, telegram_chat_id, COALESCE(is_admin, FALSE) FROM users WHERE id = $1`

	var user models.User

	err := GlobalDB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Balance,
		&user.BalanceRub,
		&user.BalanceArbitrage,
		&user.BalancePersonal,
		&user.KYCStatus,
		&user.ActiveMode,
		&user.CreatedAt,
		&user.Status,
		&user.TelegramChatID,
		&user.IsAdmin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("пользователь не найден")
		}
		log.Printf("DB Error GetUserByID: %v", err)
		return models.User{}, err
	}

	return user, nil
}

// UpdateTelegramChatID - Обновляет TelegramChatID пользователя.
// Принимает int, так как ChatID в Go - это int, а в БД - BIGINT.
func UpdateTelegramChatID(userID int, chatID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	_, err := GlobalDB.Exec(
		"UPDATE users SET telegram_chat_id = $1 WHERE id = $2",
		chatID, userID,
	)
	if err != nil {
		log.Printf("DB Error UpdateTelegramChatID: %v", err)
		return fmt.Errorf("не удалось обновить telegram_chat_id")
	}
	return nil
}

// DeductBalance - Списывает сумму с баланса пользователя (для оплаты выпуска карт и т.д.)
func DeductBalance(userID int, amount decimal.Decimal, details string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("сумма списания должна быть положительной")
	}

	tx, err := GlobalDB.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer tx.Rollback()

	// Check balance
	var balance decimal.Decimal
	err = tx.QueryRow("SELECT COALESCE(balance, 0) FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("не удалось получить баланс: %v", err)
	}
	if balance.LessThan(amount) {
		return fmt.Errorf("недостаточно средств (баланс: $%s, требуется: $%s)", balance.StringFixed(2), amount.StringFixed(2))
	}

	// Deduct
	_, err = tx.Exec(
		"UPDATE users SET balance_rub = COALESCE(balance_rub, 0) - $1, balance = balance - $2 WHERE id = $3",
		amount, amount, userID,
	)
	if err != nil {
		return fmt.Errorf("не удалось списать баланс: %v", err)
	}

	// Record transaction
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, amount, fee, transaction_type, status, details, executed_at)
			VALUES ($1, $2, $3, 'CARD_ISSUE_FEE', 'APPROVED', $4, $5)`,
		userID, amount, decimal.Zero, details, time.Now(),
	)
	if err != nil {
		log.Printf("DB Error Insert deduction transaction: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка фиксации: %v", err)
	}

	log.Printf("User %d: deducted $%s for: %s", userID, amount.StringFixed(2), details)
	return nil
}

// ProcessDeposit - Обрабатывает пополнение баланса пользователя и записывает транзакцию.
func ProcessDeposit(userID int, amount decimal.Decimal) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("сумма пополнения должна быть положительной")
	}

	// 1. НАЧАЛО ТРАНЗАКЦИИ
	tx, err := GlobalDB.Begin()
	if err != nil {
		log.Printf("DB Error Begin: %v", err)
		return fmt.Errorf("не удалось начать транзакцию")
	}
	defer tx.Rollback()

	// 2. Увеличение баланса пользователя в рублях (XPLR: balance_rub)
	_, err = tx.Exec(
		"UPDATE users SET balance_rub = COALESCE(balance_rub, 0) + $1, balance = balance + $2 WHERE id = $3",
		amount, amount, userID,
	)
	if err != nil {
		log.Printf("DB Error Update Balance: %v", err)
		return fmt.Errorf("не удалось обновить баланс")
	}

	// 3. Запись транзакции пополнения (FUND)
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, amount, fee, transaction_type, status, details, executed_at)
			VALUES ($1, $2, $3, 'FUND', 'APPROVED', $4, $5)`,
		userID,
		amount,
		decimal.Zero, // Комиссия
		fmt.Sprintf("Deposit via API. Amount: %s", amount.String()),
		time.Now(),
	)
	if err != nil {
		log.Printf("DB Error Insert Transaction: %v", err)
		return fmt.Errorf("не удалось записать транзакцию")
	}

	// 4. КОММИТ
	if err := tx.Commit(); err != nil {
		log.Printf("DB Error Commit: %v", err)
		return fmt.Errorf("ошибка фиксации транзакции")
	}

	log.Printf("User %d successfully deposited %s. Transaction committed.", userID, amount.String())

	// 5. ОТПРАВКА TELEGRAM УВЕДОМЛЕНИЯ
	user, err := GetUserByID(userID)
	if err == nil && user.TelegramChatID.Valid {
		message := fmt.Sprintf("✅ Deposit Successful!\n\nAmount: %s ₽\nNew Balance: %s ₽",
			amount.String(),
			user.BalanceRub.String())
		notification.SendTelegramMessage(user.TelegramChatID.Int64, message)
	}

	return nil
}
