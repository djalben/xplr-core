package postgres

import (
	"context"

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

// GetByID — получение пользователя.
func (r *userRepo) GetByID(ctx context.Context, id domain.UUID) (*domain.User, error) {
	const query = `SELECT id, email, password_hash, kyc_status, status, telegram_chat_id, 
	                      referral_code, referred_by, created_at 
	               FROM users WHERE id = $1`

	var u domain.User

	err := r.store.GetContext(ctx, &u, query, id)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &u, nil
}

// GetByEmail — по email.
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `SELECT id, email, password_hash, kyc_status, status, telegram_chat_id, 
	                      referral_code, referred_by, created_at 
	               FROM users WHERE email = $1`

	var u domain.User

	err := r.store.GetContext(ctx, &u, query, email)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &u, nil
}

// Save — создание пользователя.
func (r *userRepo) Save(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (id, email, password_hash, kyc_status, status, 
		                   telegram_chat_id, referral_code, referred_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.store.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.KYCStatus,
		user.Status, user.TelegramChatID, user.ReferralCode,
		user.ReferredBy, user.CreatedAt,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// Update — обновление пользователя.
func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	const query = `
		UPDATE users 
		SET kyc_status = $1, status = $2, telegram_chat_id = $3, 
		    referral_code = $4, referred_by = $5 
		WHERE id = $6`

	_, err := r.store.ExecContext(ctx, query,
		user.KYCStatus, user.Status, user.TelegramChatID,
		user.ReferralCode, user.ReferredBy, user.ID,
	)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
