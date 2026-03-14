package app

import (
	"context"

	"github.com/djalben/xplr-core/internal/application/card"
	"github.com/djalben/xplr-core/internal/application/commission"
	"github.com/djalben/xplr-core/internal/application/ticket"
	"github.com/djalben/xplr-core/internal/application/transaction"
	"github.com/djalben/xplr-core/internal/application/wallet"
	"github.com/djalben/xplr-core/internal/config"
	"github.com/djalben/xplr-core/internal/infrastructure/persistence/postgres"
	"github.com/djalben/xplr-core/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// Container — главный DI-контейнер.
type Container struct {
	DB *sqlx.DB

	WalletRepo      ports.WalletRepository
	CardRepo        ports.CardRepository
	TransactionRepo ports.TransactionRepository
	TicketRepo      ports.TicketRepository
	UserRepo        ports.UserRepository
	CommissionRepo  ports.CommissionConfigRepository

	WalletUseCase      *wallet.UseCase
	CardUseCase        *card.UseCase
	TransactionUseCase *transaction.UseCase
	TicketUseCase      *ticket.UseCase
	CommissionUseCase  *commission.UseCase
}

// NewContainer — создаёт всё приложение.
func NewContainer(cfg *config.ENV) (*Container, error) { // ← исправлено на ENV
	ctx := context.Background()

	// Подключаемся к БД
	db, err := postgres.Connect(ctx, cfg.PostgresDSN) // ← исправлено
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	// Репозитории
	walletRepo := postgres.NewWalletRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	ticketRepo := postgres.NewTicketRepository(db)
	userRepo := postgres.NewUserRepository(db)
	commissionRepo := postgres.NewCommissionConfigRepository(db)

	return &Container{
		DB: db,

		WalletRepo:      walletRepo,
		CardRepo:        cardRepo,
		TransactionRepo: transactionRepo,
		TicketRepo:      ticketRepo,
		UserRepo:        userRepo,
		CommissionRepo:  commissionRepo,

		WalletUseCase:      wallet.NewUseCase(walletRepo, transactionRepo),
		CardUseCase:        card.NewUseCase(cardRepo, walletRepo, transactionRepo),
		TransactionUseCase: transaction.NewUseCase(transactionRepo),
		TicketUseCase:      ticket.NewUseCase(ticketRepo),
		CommissionUseCase:  commission.NewUseCase(commissionRepo),
	}, nil
}

// Close — корректное завершение.
func (c *Container) Close() error {
	if c.DB != nil {
		return wrapper.Wrap(c.DB.Close())
	}

	return nil
}
