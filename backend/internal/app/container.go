package app

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/application/commission"
	"github.com/djalben/xplr-core/backend/internal/application/grades"
	"github.com/djalben/xplr-core/backend/internal/application/ticket"
	"github.com/djalben/xplr-core/backend/internal/application/transaction"
	"github.com/djalben/xplr-core/backend/internal/application/wallet"
	"github.com/djalben/xplr-core/backend/internal/config"
	"github.com/djalben/xplr-core/backend/internal/infrastructure/persistence/postgres"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Container struct {
	DB *sqlx.DB

	WalletRepo      ports.WalletRepository
	CardRepo        ports.CardRepository
	TransactionRepo ports.TransactionRepository
	TicketRepo      ports.TicketRepository
	UserRepo        ports.UserRepository
	CommissionRepo  ports.CommissionConfigRepository
	GradeRepo       ports.GradeRepository

	WalletUseCase      *wallet.UseCase
	CardUseCase        *card.UseCase
	TransactionUseCase *transaction.UseCase
	TicketUseCase      *ticket.UseCase
	CommissionUseCase  *commission.UseCase
	GradesUseCase      *grades.UseCase
}

func NewContainer(cfg *config.ENV) (*Container, error) {
	ctx := context.Background()

	db, err := postgres.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	walletRepo := postgres.NewWalletRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	ticketRepo := postgres.NewTicketRepository(db)
	userRepo := postgres.NewUserRepository(db)
	commissionRepo := postgres.NewCommissionConfigRepository(db)
	gradeRepo := postgres.NewGradeRepository(db)

	// WalletUseCase создаём первым — он нужен для CardUseCase
	walletUC := wallet.NewUseCase(walletRepo, transactionRepo)

	return &Container{
		DB: db,

		WalletRepo:      walletRepo,
		CardRepo:        cardRepo,
		TransactionRepo: transactionRepo,
		TicketRepo:      ticketRepo,
		UserRepo:        userRepo,
		CommissionRepo:  commissionRepo,
		GradeRepo:       gradeRepo,

		WalletUseCase:      walletUC,
		CardUseCase:        card.NewUseCase(cardRepo, walletRepo, transactionRepo, walletUC),
		TransactionUseCase: transaction.NewUseCase(transactionRepo),
		TicketUseCase:      ticket.NewUseCase(ticketRepo),
		CommissionUseCase:  commission.NewUseCase(commissionRepo),
		GradesUseCase:      grades.NewUseCase(gradeRepo),
	}, nil
}

func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
