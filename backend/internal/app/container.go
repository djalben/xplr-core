package app

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/application/auth"
	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/application/commission"
	"github.com/djalben/xplr-core/backend/internal/application/grades"
	"github.com/djalben/xplr-core/backend/internal/application/kyc"
	"github.com/djalben/xplr-core/backend/internal/application/store"
	"github.com/djalben/xplr-core/backend/internal/application/ticket"
	"github.com/djalben/xplr-core/backend/internal/application/transaction"
	"github.com/djalben/xplr-core/backend/internal/application/user"
	"github.com/djalben/xplr-core/backend/internal/application/wallet"
	"github.com/djalben/xplr-core/backend/internal/config"
	"github.com/djalben/xplr-core/backend/internal/infrastructure/mailer"
	"github.com/djalben/xplr-core/backend/internal/infrastructure/persistence/postgres"
	esimProvider "github.com/djalben/xplr-core/backend/internal/infrastructure/providers/esim"
	vpnProvider "github.com/djalben/xplr-core/backend/internal/infrastructure/providers/vpn"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Container struct {
	DB *sqlx.DB

	WalletRepo        ports.WalletRepository
	CardRepo          ports.CardRepository
	TransactionRepo   ports.TransactionRepository
	TicketRepo        ports.TicketRepository
	UserRepo          ports.UserRepository
	TrustedDeviceRepo ports.TrustedDeviceRepository
	AuthRateLimiter   ports.AuthRateLimiter
	ReferralRepo      ports.ReferralRepository
	CommissionRepo    ports.CommissionConfigRepository
	GradeRepo         ports.GradeRepository
	ExchangeRateRepo  ports.ExchangeRateRepository
	KYCRepo           ports.KYCApplicationRepository
	StoreRepo         ports.StoreRepository
	NewsRepo          ports.NewsRepository
	SystemRepo        ports.SystemSettingsRepository
	AdminLogsRepo     ports.AdminLogsRepository
	AdminDashRepo     ports.AdminDashboardRepository

	TelegramBotUsername string

	AuthUseCase        *auth.UseCase
	UserUseCase        *user.UseCase
	WalletUseCase      *wallet.UseCase
	CardUseCase        *card.UseCase
	TransactionUseCase *transaction.UseCase
	TicketUseCase      *ticket.UseCase
	CommissionUseCase  *commission.UseCase
	GradesUseCase      *grades.UseCase
	KYCUseCase         *kyc.UseCase
	StoreUseCase       *store.UseCase

	VPNAdminProvider ports.VPNAdminProvider
}

func NewContainer(ctx context.Context, cfg *config.ENV) (*Container, error) {
	db, err := postgres.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	walletRepo := postgres.NewWalletRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	ticketRepo := postgres.NewTicketRepository(db)
	userRepo := postgres.NewUserRepository(db)
	trustedDeviceRepo := postgres.NewTrustedDeviceRepository(db)
	authLimiter := postgres.NewAuthRateLimiter(db)
	commissionRepo := postgres.NewCommissionConfigRepository(db)
	gradeRepo := postgres.NewGradeRepository(db)
	exchangeRateRepo := postgres.NewExchangeRateRepository(db)
	referralRepo := postgres.NewReferralRepository(db)
	kycRepo := postgres.NewKYCApplicationRepository(db)
	storeRepo := postgres.NewStoreRepository(db)
	newsRepo := postgres.NewNewsRepository(db)
	systemRepo := postgres.NewSystemSettingsRepository(db)
	adminLogsRepo := postgres.NewAdminLogsRepository(db)
	adminDashRepo := postgres.NewAdminDashboardRepository(db)

	var mail ports.Mailer = mailer.Noop{}
	if cfg.SMTPHost != "" {
		mail = &mailer.SMTP{
			Host: cfg.SMTPHost, Port: cfg.SMTPPort,
			User: cfg.SMTPUser, Password: cfg.SMTPPassword,
			From: cfg.SMTPFrom,
		}
	}

	// WalletUseCase создаём первым — он нужен для CardUseCase
	walletUC := wallet.NewUseCase(walletRepo, transactionRepo, systemRepo)
	authUC := auth.NewUseCase(
		userRepo,
		walletRepo,
		gradeRepo,
		trustedDeviceRepo,
		[]byte(cfg.JWTSecret),
		mail,
		cfg.AppPublicURL,
	)
	userUC := user.NewUseCase(userRepo, walletRepo, gradeRepo, referralRepo, commissionRepo, systemRepo)
	kycUC := kyc.NewUseCase(kycRepo, userRepo)
	esim := esimProvider.NewMobiMatterProvider(*cfg)
	vpn := vpnProvider.NewVlessXPanelProvider(*cfg)
	var vpnPort ports.VPNProvider
	var vpnAdmin ports.VPNAdminProvider
	if vpn.Enabled() {
		vpnPort = vpn
		vpnAdmin = vpn
	}
	storeUC := store.NewUseCase(storeRepo, walletRepo, esim, vpnPort)

	return &Container{
		DB: db,

		WalletRepo:        walletRepo,
		CardRepo:          cardRepo,
		TransactionRepo:   transactionRepo,
		TicketRepo:        ticketRepo,
		UserRepo:          userRepo,
		TrustedDeviceRepo: trustedDeviceRepo,
		AuthRateLimiter:   authLimiter,
		ReferralRepo:      referralRepo,
		CommissionRepo:    commissionRepo,
		GradeRepo:         gradeRepo,
		ExchangeRateRepo:  exchangeRateRepo,
		KYCRepo:           kycRepo,
		StoreRepo:         storeRepo,
		NewsRepo:          newsRepo,
		SystemRepo:        systemRepo,
		AdminLogsRepo:     adminLogsRepo,
		AdminDashRepo:     adminDashRepo,

		TelegramBotUsername: cfg.TelegramBotUsername,

		AuthUseCase:        authUC,
		UserUseCase:        userUC,
		WalletUseCase:      walletUC,
		CardUseCase:        card.NewUseCase(cardRepo, walletRepo, transactionRepo, gradeRepo, walletUC),
		TransactionUseCase: transaction.NewUseCase(transactionRepo),
		TicketUseCase:      ticket.NewUseCase(ticketRepo),
		CommissionUseCase:  commission.NewUseCase(commissionRepo),
		GradesUseCase:      grades.NewUseCase(gradeRepo),
		KYCUseCase:         kycUC,
		StoreUseCase:       storeUC,
		VPNAdminProvider:   vpnAdmin,
	}, nil
}

func (c *Container) Close() error {
	if c.DB != nil {
		return wrapper.Wrap(c.DB.Close())
	}

	return nil
}
