package domain

import (
	"time"
)

type User struct {
	ID             UUID       `json:"id" db:"id"`
	Email          string     `json:"email" db:"email"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	IsAdmin        bool       `json:"isAdmin" db:"is_admin"`
	KYCStatus      KYCStatus  `json:"kycStatus" db:"kyc_status"`
	Status         UserStatus `json:"status" db:"status"`
	TelegramChatID *int64     `json:"telegramChatId,omitempty" db:"telegram_chat_id"`
	ReferralCode   string     `json:"referralCode" db:"referral_code"`
	ReferredBy     *UUID      `json:"referredBy,omitempty" db:"referred_by"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`

	EmailVerified          bool       `json:"emailVerified" db:"email_verified"`
	EmailVerifyTokenHash   *string    `json:"-" db:"email_verify_token_hash"`
	EmailVerifyExpiresAt   *time.Time `json:"-" db:"email_verify_expires_at"`
	PasswordResetTokenHash *string    `json:"-" db:"password_reset_token_hash"`
	PasswordResetExpiresAt *time.Time `json:"-" db:"password_reset_expires_at"`
	TOTPSecret             *string    `json:"-" db:"totp_secret"`
	TOTPEnabled            bool       `json:"totpEnabled" db:"totp_enabled"`
	NotifyEmail            bool       `json:"notifyEmail" db:"notify_email"`
	NotifyTelegram         bool       `json:"notifyTelegram" db:"notify_telegram"`
	NotifyTransactions     bool       `json:"notifyTransactions" db:"notify_transactions"`
	NotifyBalance          bool       `json:"notifyBalance" db:"notify_balance"`
	NotifySecurity         bool       `json:"notifySecurity" db:"notify_security"`
	NotifyCardOperations   bool       `json:"notifyCardOperations" db:"notify_card_operations"`
	TelegramLinkCode       *string    `json:"-" db:"telegram_link_code"`
	TelegramLinkExpiresAt  *time.Time `json:"-" db:"telegram_link_expires_at"`
}

type (
	UserStatus string
	KYCStatus  string
)

const (
	UserStatusActive  UserStatus = "ACTIVE"
	UserStatusBlocked UserStatus = "BLOCKED"

	KYCPending  KYCStatus = "PENDING"
	KYCApproved KYCStatus = "APPROVED"
	KYCRejected KYCStatus = "REJECTED"
)

func NewUser(email, passwordHash string) (*User, error) {
	if email == "" {
		return nil, NewInvalidInput("email is required")
	}

	return &User{
		ID:                   NewUUID(),
		Email:                email,
		PasswordHash:         passwordHash,
		KYCStatus:            KYCPending,
		Status:               UserStatusActive,
		EmailVerified:        false,
		NotifyEmail:          true,
		NotifyTelegram:       true,
		NotifyTransactions:   true,
		NotifyBalance:        true,
		NotifySecurity:       true,
		NotifyCardOperations: true,
		CreatedAt:            time.Now().UTC(),
	}, nil
}
