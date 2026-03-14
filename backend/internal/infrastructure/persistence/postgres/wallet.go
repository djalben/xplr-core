package postgres

import (
	"context"

	"github.com/djalben/xplr-core/internal/domain"
	"github.com/djalben/xplr-core/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type walletRepo struct {
	store *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) ports.WalletRepository {
	return &walletRepo{store: db}
}

func (r *walletRepo) GetByUserID(ctx context.Context, userID domain.UUID) (*domain.Wallet, error) {
	const query = `SELECT id, user_id, balance, created_at FROM wallets WHERE user_id = $1`

	var w domain.Wallet

	err := r.store.GetContext(ctx, &w, query, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &w, nil
}

func (r *walletRepo) Update(ctx context.Context, wallet *domain.Wallet) error {
	const query = `UPDATE wallets SET balance = $1 WHERE id = $2`

	_, err := r.store.ExecContext(ctx, query, wallet.Balance, wallet.ID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
