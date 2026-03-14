package postgres

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type walletRepo struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) ports.WalletRepository {
	return &walletRepo{db: db}
}

func (r *walletRepo) GetByUserID(ctx context.Context, userID domain.UUID) (*domain.Wallet, error) {
	const query = `
		SELECT id, user_id, balance, auto_topup_enabled, created_at 
		FROM wallets 
		WHERE user_id = $1`

	var w domain.Wallet

	err := r.db.GetContext(ctx, &w, query, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &w, nil
}

func (r *walletRepo) Update(ctx context.Context, wallet *domain.Wallet) error {
	const query = `
		UPDATE wallets 
		SET balance = $1, 
		    auto_topup_enabled = $2 
		WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, wallet.Balance, wallet.AutoTopUpEnabled, wallet.ID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}