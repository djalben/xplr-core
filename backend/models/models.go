package models

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// --- СТРУКТУРЫ ЗАПРОСОВ К API (НОВЫЙ БЛОК) ---

// RegisterRequest - Запрос на регистрацию пользователя
type RegisterRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	ReferralCode string `json:"referral_code,omitempty"` // Опционально
}

// LoginRequest - Запрос на вход пользователя
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// FundRequest - Запрос на пополнение баланса (используется в deposit.go)
type FundRequest struct {
	Amount decimal.Decimal `json:"amount"`
}

// AuthRequest - Запрос на авторизацию карты (используется в handlers/test_authorize.go)
type AuthRequest struct {
	CardID      int             `json:"card_id"`
	Amount      decimal.Decimal `json:"amount"`
	MerchantName string         `json:"merchant_name"`
}

// --- СТРУКТУРЫ ПОЛЬЗОВАТЕЛЕЙ И АУТЕНТИФИКАЦИИ ---

// User - Структура для пользователя
// В Supabase: id (UUID), email, password_hash, balance_rub (numeric), active_mode
// Для совместимости с кодом используем int для ID (конвертация UUID -> int при необходимости)
type User struct {
	ID             int             `json:"id"`              // В Supabase: UUID (конвертируется)
	Email          string          `json:"email"`
	PasswordHash   string          `json:"-"`               // password_hash в Supabase
	Balance        decimal.Decimal `json:"balance"`         // Legacy поле
	BalanceRub     decimal.Decimal `json:"balance_rub"`     // Основной баланс в рублях (XPLR) - соответствует Supabase
	KYCStatus     string          `json:"kyc_status"`      // Статус верификации (например: pending, verified, rejected)
	ActiveMode     string          `json:"active_mode"`     // Режим работы: по умолчанию 'personal' - соответствует Supabase
	CreatedAt      time.Time       `json:"created_at"`
	Status         string          `json:"status"`
	TeamID         sql.NullInt64   `json:"team_id"`
	TelegramChatID sql.NullInt64   `json:"telegram_chat_id"`
}

// APIKey - Структура для ключей
type APIKey struct {
	Key         string    `json:"api_key"`
	UserID      int       `json:"user_id"`
	Permissions string    `json:"permissions"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// DepositRequest - Структура для запроса пополнения (используется API Key)
type DepositRequest struct {
	Amount decimal.Decimal `json:"amount"`
	UserID int             `json:"user_id"`
}

// --- СТРУКТУРЫ ТРАНЗАКЦИЙ И ОТЧЕТНОСТИ ---

// Transaction - Структура для транзакции
type Transaction struct {
	TransactionID    int             `json:"transaction_id"`
	UserID           int             `json:"user_id"`
	UserEmail        string          `json:"user_email,omitempty"`
	CardID           *int            `json:"card_id,omitempty"`
	CardLast4Digits  string          `json:"card_last_4_digits,omitempty"`
	Amount           decimal.Decimal `json:"amount"`
	Fee              decimal.Decimal `json:"fee"`
	TransactionType  string          `json:"transaction_type"`
	Status           string          `json:"status"`
	Details          string          `json:"details"`
	ProviderTxID     string          `json:"provider_tx_id,omitempty"` // ID транзакции от провайдера (Wallester) для idempotency
	ExecutedAt       time.Time       `json:"executed_at"`
}

// --- СТРУКТУРЫ АВТОРИЗАЦИИ ---

// AuthResponse - Ответ от логики authorizeCard
type AuthResponse struct {
	Success bool            `json:"success"`
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Fee     decimal.Decimal `json:"fee"`
}

// --- СТРУКТУРЫ УПРАВЛЕНИЯ КАРТАМИ ---

// Card - Структура для карты
// В Supabase: id (UUID), service_id (int), external_id, bin, last_4, status
// Для совместимости сохраняем текущие поля, добавляем соответствие Supabase полям
type Card struct {
	ID                    int             `json:"id"`                    // В Supabase: UUID (конвертируется)
	UserID                int             `json:"user_id"`              // В Supabase: UUID user_id (конвертируется)
	TeamID                *int            `json:"team_id,omitempty"`
	ProviderCardID        string          `json:"provider_card_id"`      // Соответствует external_id в Supabase
	ExternalID            string          `json:"external_id,omitempty"` // Поле из Supabase (дублирует ProviderCardID)
	BIN                   string          `json:"bin"`                   // Соответствует Supabase bin
	Last4Digits           string          `json:"last_4_digits"`         // Соответствует Supabase last_4
	CardStatus            string          `json:"card_status"`           // Соответствует Supabase status
	Status                string          `json:"status,omitempty"`      // Прямое соответствие Supabase status (дублирует CardStatus)
	Nickname              string          `json:"nickname"`
	ServiceSlug           string          `json:"service_slug"`          // Метка: 'arbitrage', 'travel', 'subscriptions' (для API)
	ServiceID             *int            `json:"service_id,omitempty"` // ID из таблицы services в Supabase
	DailySpendLimit       decimal.Decimal `json:"daily_spend_limit"`
	FailedAuthCount       int             `json:"failed_auth_count"`
	CardType              string          `json:"card_type"`
	AutoReplenishEnabled  bool            `json:"auto_replenish_enabled"`
	AutoReplenishThreshold decimal.Decimal `json:"auto_replenish_threshold"`
	AutoReplenishAmount   decimal.Decimal `json:"auto_replenish_amount"`
	CardBalance           decimal.Decimal `json:"card_balance"`         // Виртуальный баланс: при списке карт = BalanceRub пользователя
	CreatedAt             time.Time       `json:"created_at"`
}

// MassIssueRequest - Запрос на массовый выпуск карт
type MassIssueRequest struct {
	Count             int             `json:"count"`
	DailyLimit        decimal.Decimal `json:"daily_limit"`
	CardNickname      string          `json:"nickname"`
	MerchantName      string          `json:"merchant_name"`
	CardType          string          `json:"card_type"`   // VISA или MasterCard
	ServiceSlug       string          `json:"service_slug"` // 'arbitrage', 'travel', 'subscriptions'
	TeamID            *int            `json:"team_id,omitempty"`
}

// CardIssueResult - Результат выпуска одной карты 
type CardIssueResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Card      *Card  `json:"card,omitempty"`
	Status    string `json:"status"`    
	CardLast4 string `json:"card_last_4"` 
	Nickname  string `json:"nickname"`  
}

// MassIssueResponse - Ответ на массовый выпуск карт 
type MassIssueResponse struct {
	Successful int               `json:"successful_count"` 
	Failed     int               `json:"failed_count"`     
	Results    []CardIssueResult `json:"results"`
}

// --- СТРУКТУРЫ АВТОПОПОЛНЕНИЯ КАРТ ---

// AutoReplenishRequest - Запрос на настройку автопополнения
type AutoReplenishRequest struct {
	Enabled  bool            `json:"enabled"`
	Threshold decimal.Decimal `json:"threshold"`
	Amount   decimal.Decimal `json:"amount"`
}

// --- СТРУКТУРЫ КОМАНД ---

// Team - Команда
type Team struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TeamMember - Участник команды
type TeamMember struct {
	ID        int       `json:"id"`
	TeamID    int       `json:"team_id"`
	UserID    int       `json:"user_id"`
	Role      string    `json:"role"` // 'owner', 'admin', 'member'
	InvitedBy *int      `json:"invited_by,omitempty"`
	JoinedAt  time.Time `json:"joined_at"`
	User      *User     `json:"user,omitempty"` // Для деталей пользователя
}

// CreateTeamRequest - Запрос на создание команды
type CreateTeamRequest struct {
	Name string `json:"name"`
}

// InviteTeamMemberRequest - Запрос на приглашение участника
type InviteTeamMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"` // 'admin' или 'member'
}

// UpdateTeamMemberRoleRequest - Запрос на изменение роли
type UpdateTeamMemberRoleRequest struct {
	Role string `json:"role"`
}

// --- СТРУКТУРЫ GRADE СИСТЕМЫ ---

// UserGrade - Grade пользователя
type UserGrade struct {
	ID         int             `json:"id"`
	UserID     int             `json:"user_id"`
	Grade      string          `json:"grade"` // 'STANDARD', 'SILVER', 'GOLD', 'PLATINUM', 'BLACK'
	TotalSpent decimal.Decimal `json:"total_spent"`
	FeePercent decimal.Decimal `json:"fee_percent"` // Комиссия в процентах (6.70 = 6.7%)
	UpdatedAt  time.Time       `json:"updated_at"`
}

// GradeInfo - Информация о Grade для отображения
type GradeInfo struct {
	Grade      string          `json:"grade"`
	TotalSpent decimal.Decimal `json:"total_spent"`
	FeePercent decimal.Decimal `json:"fee_percent"`
	NextGrade  *string         `json:"next_grade,omitempty"`
	NextSpend  *decimal.Decimal `json:"next_spend,omitempty"` // Сколько нужно потратить до следующего уровня
}

// --- СТРУКТУРЫ РЕФЕРАЛЬНОЙ ПРОГРАММЫ ---

// Referral - Реферал
type Referral struct {
	ID              int             `json:"id"`
	ReferrerID      int             `json:"referrer_id"`
	ReferredID      int             `json:"referred_id"`
	ReferralCode    string          `json:"referral_code"`
	Status          string          `json:"status"` // 'PENDING', 'ACTIVE', 'COMPLETED'
	CommissionEarned decimal.Decimal `json:"commission_earned"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ReferralStats - Статистика реферальной программы
type ReferralStats struct {
	TotalReferrals    int             `json:"total_referrals"`
	ActiveReferrals   int             `json:"active_referrals"`
	TotalCommission   decimal.Decimal `json:"total_commission"`
	ReferralCode      string          `json:"referral_code"`
}