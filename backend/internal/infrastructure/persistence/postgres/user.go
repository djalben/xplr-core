package postgres

import (
	"context"
	"strings"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type userRepo struct {
	store *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) ports.UserRepository {
	return &userRepo{store: db}
}

const userSelectColumns = `id, email, password_hash, is_admin, kyc_status, status, telegram_chat_id,
	referral_code, referred_by, created_at,
	email_verified, email_verify_token_hash, email_verify_expires_at,
	password_reset_token_hash, password_reset_expires_at,
	totp_secret, totp_enabled, notify_email, notify_telegram,
	notify_transactions, notify_balance, notify_security, notify_card_operations,
	telegram_link_code, telegram_link_expires_at,
	last_read_news_at, news_notifications_enabled`

// GetByID — получение пользователя.
func (r *userRepo) GetByID(ctx context.Context, id domain.UUID) (*domain.User, error) {
	const query = `SELECT ` + userSelectColumns + ` FROM users WHERE id = $1`

	var u domain.User

	err := r.store.GetContext(ctx, &u, query, id)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &u, nil
}

// GetByEmail — по email.
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `SELECT ` + userSelectColumns + ` FROM users WHERE email = $1`

	var u domain.User

	err := r.store.GetContext(ctx, &u, query, email)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &u, nil
}

// GetByEmailVerifyTokenHash — по хэшу токена верификации email.
func (r *userRepo) GetByEmailVerifyTokenHash(ctx context.Context, tokenHash string) (*domain.User, error) {
	const query = `SELECT ` + userSelectColumns + ` FROM users WHERE email_verify_token_hash = $1`

	var u domain.User

	err := r.store.GetContext(ctx, &u, query, tokenHash)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &u, nil
}

// GetByPasswordResetTokenHash — по хэшу токена сброса пароля.
func (r *userRepo) GetByPasswordResetTokenHash(ctx context.Context, tokenHash string) (*domain.User, error) {
	const query = `SELECT ` + userSelectColumns + ` FROM users WHERE password_reset_token_hash = $1`

	var u domain.User

	err := r.store.GetContext(ctx, &u, query, tokenHash)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &u, nil
}

// GetByTelegramChatID — пользователь с привязанным чатом Telegram.
func (r *userRepo) GetByTelegramChatID(ctx context.Context, chatID int64) (*domain.User, error) {
	const query = `SELECT ` + userSelectColumns + ` FROM users WHERE telegram_chat_id = $1`

	var u domain.User

	err := r.store.GetContext(ctx, &u, query, chatID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &u, nil
}

// Save — создание пользователя.
func (r *userRepo) Save(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (
			id, email, password_hash, is_admin, kyc_status, status,
			telegram_chat_id, referral_code, referred_by, created_at,
			email_verified, email_verify_token_hash, email_verify_expires_at,
			password_reset_token_hash, password_reset_expires_at,
			totp_secret, totp_enabled, notify_email, notify_telegram,
			notify_transactions, notify_balance, notify_security, notify_card_operations,
			telegram_link_code, telegram_link_expires_at,
			last_read_news_at, news_notifications_enabled
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27
		)`

	_, err := r.store.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.IsAdmin, user.KYCStatus,
		user.Status, user.TelegramChatID, user.ReferralCode,
		user.ReferredBy, user.CreatedAt,
		user.EmailVerified, user.EmailVerifyTokenHash, user.EmailVerifyExpiresAt,
		user.PasswordResetTokenHash, user.PasswordResetExpiresAt,
		user.TOTPSecret, user.TOTPEnabled, user.NotifyEmail, user.NotifyTelegram,
		user.NotifyTransactions, user.NotifyBalance, user.NotifySecurity, user.NotifyCardOperations,
		user.TelegramLinkCode, user.TelegramLinkExpiresAt,
		user.LastReadNewsAt, user.NewsNotificationsEnabled,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// Update — полное обновление изменяемых полей пользователя.
func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	const query = `
		UPDATE users SET
			password_hash = $1,
			kyc_status = $2, status = $3, telegram_chat_id = $4,
			referral_code = $5, referred_by = $6,
			email_verified = $7, email_verify_token_hash = $8, email_verify_expires_at = $9,
			password_reset_token_hash = $10, password_reset_expires_at = $11,
			totp_secret = $12, totp_enabled = $13, notify_email = $14, notify_telegram = $15,
			notify_transactions = $16, notify_balance = $17, notify_security = $18, notify_card_operations = $19,
			telegram_link_code = $20, telegram_link_expires_at = $21,
			last_read_news_at = $22, news_notifications_enabled = $23
		WHERE id = $24`

	_, err := r.store.ExecContext(ctx, query,
		user.PasswordHash, user.KYCStatus, user.Status, user.TelegramChatID,
		user.ReferralCode, user.ReferredBy,
		user.EmailVerified, user.EmailVerifyTokenHash, user.EmailVerifyExpiresAt,
		user.PasswordResetTokenHash, user.PasswordResetExpiresAt,
		user.TOTPSecret, user.TOTPEnabled, user.NotifyEmail, user.NotifyTelegram,
		user.NotifyTransactions, user.NotifyBalance, user.NotifySecurity, user.NotifyCardOperations,
		user.TelegramLinkCode, user.TelegramLinkExpiresAt,
		user.LastReadNewsAt, user.NewsNotificationsEnabled,
		user.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "23505") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return wrapper.Wrap(domain.NewAlreadyExists("telegram chat is already linked to another account"))
		}

		return wrapper.Wrap(err)
	}

	return nil
}

// SearchByEmail — поиск по подстроке email (ILIKE), новее первыми.
func (r *userRepo) SearchByEmail(ctx context.Context, query string, limit int) ([]*domain.User, error) {
	if limit <= 0 {
		limit = 50
	}

	const q = `SELECT ` + userSelectColumns + ` FROM users WHERE email ILIKE $1 ORDER BY created_at DESC LIMIT $2`

	pattern := "%" + query + "%"

	var out []*domain.User

	err := r.store.SelectContext(ctx, &out, q, pattern, limit)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return out, nil
}

// ListUsers — постраничный список по дате регистрации (новее первыми).
func (r *userRepo) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	if limit <= 0 {
		limit = 50
	}

	const q = `SELECT ` + userSelectColumns + ` FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	var out []*domain.User

	err := r.store.SelectContext(ctx, &out, q, limit, offset)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return out, nil
}

// CountUsers — число записей в users.
func (r *userRepo) CountUsers(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(*) FROM users`

	var n int64

	err := r.store.GetContext(ctx, &n, q)
	if err != nil {
		return 0, wrapper.Wrap(err)
	}

	return n, nil
}

// SetUserStatus — смена статуса аккаунта (ACTIVE/BLOCKED).
func (r *userRepo) SetUserStatus(ctx context.Context, id domain.UUID, status domain.UserStatus) error {
	const q = `UPDATE users SET status = $1 WHERE id = $2`

	_, err := r.store.ExecContext(ctx, q, status, id)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// SetIsAdmin — выдача/снятие флага администратора.
func (r *userRepo) SetIsAdmin(ctx context.Context, id domain.UUID, isAdmin bool) error {
	const q = `UPDATE users SET is_admin = $1 WHERE id = $2`

	_, err := r.store.ExecContext(ctx, q, isAdmin, id)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
