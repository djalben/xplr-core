package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"

	// ВАЖНО: Убедитесь, что пути верны
	"github.com/djalben/xplr-core/backend/handler"
	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/djalben/xplr-core/backend/telegram"
	"github.com/djalben/xplr-core/backend/usecase"
)

// DB - глобальная переменная для подключения к базе данных (будет использоваться только здесь)
var DB *sql.DB

func main() {
	// 0. Загрузка .env файла (если существует)
	// Try multiple .env locations depending on where binary is run from
	_ = godotenv.Load("backend/.env")
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load(".env")

	// 1. Инициализация базы данных
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	var err error
	DB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer DB.Close()

	// Проверка соединения
	if err = DB.PingContext(context.Background()); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Передача DB в пакет handlers
	handler.GlobalDB = DB
	// Передача DB в пакет repository (КРИТИЧЕСКИ ВАЖНО!)
	repository.GlobalDB = DB

	log.Println("Successfully connected to the database!")

	// 1.1. Schema Integrity Guard — проверяет и создаёт отсутствующие колонки
	repository.RunSchemaGuard()

	// 1.2. Shop infrastructure — fulfillment engine + deposit monitor
	handler.InitShopInfrastructure()

	// Start Aeza hosting balance monitor (alerts admins when balance < threshold)
	service.StartAezaBalanceMonitor()

	// Тест "Дыхания" — проверка таблицы services в Supabase
	log.Println("Testing Supabase connection: SELECT slug FROM services...")
	rows, err := DB.Query("SELECT slug FROM services")
	if err != nil {
		log.Printf("⚠️  Warning: Could not query services table: %v", err)
		log.Println("   (This is OK if the table doesn't exist yet or schema differs)")
	} else {
		defer rows.Close()
		var slugs []string
		for rows.Next() {
			var slug string
			if err := rows.Scan(&slug); err != nil {
				log.Printf("Error scanning slug: %v", err)
				continue
			}
			slugs = append(slugs, slug)
		}
		if err := rows.Err(); err != nil {
			log.Printf("Error iterating services: %v", err)
		} else {
			log.Println("✅ Supabase connection verified!")
			if len(slugs) > 0 {
				fmt.Println("   Services found:", slugs)
			} else {
				fmt.Println("   Services table is empty (expected: arbitrage, travel, subscriptions)")
			}
		}
	}

	// Ensure chat tables exist
	if err := repository.EnsureChatTables(); err != nil {
		log.Printf("⚠️ Warning: could not ensure chat tables: %v", err)
	}

	// Telegram bot token (для реальной отправки уведомлений)
	// CRITICAL: Сервер НЕ запустится без токена — уведомления обязательны
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if tgToken == "" {
		log.Fatal("🚨 [FATAL] TELEGRAM_BOT_TOKEN is EMPTY — server cannot start without notification service. Set the env var and restart.")
	}
	telegram.SetBotToken(tgToken)
	telegram.AdminChatIDsProvider = repository.GetAdminChatIDs
	log.Printf("✅ [INIT] Telegram bot token set (%d chars): real notifications enabled", len(tgToken))

	// SMTP check — fatal if not configured
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	if smtpHost == "" || smtpPort == "" {
		log.Fatal("🚨 [FATAL] SMTP_HOST/SMTP_PORT not set — server cannot start without email notification service.")
	}
	if smtpUser == "" || smtpPass == "" {
		log.Fatal("🚨 [FATAL] SMTP_USER/SMTP_PASS not set — server cannot start without email credentials.")
	}
	log.Printf("✅ [INIT] SMTP configured: host=%s, port=%s, user=%s", smtpHost, smtpPort, smtpUser)

	// Initialize card provider (MockProvider by default, ArmeniaProvider if configured)
	service.InitCardProvider(repository.GlobalDB)
	log.Printf("✅ [INIT] Card provider: %s", service.GetCardProvider().GetProviderName())

	// 1.5. Запуск фонового процесса автопополнения карт
	go usecase.StartAutoReplenishmentWorker()

	// 1.7. Запуск cron-задачи: возврат остатков истёкших карт в Кошелёк
	go usecase.StartExpiryReclaimWorker()

	// REMOVED: Wallester balance sync - provider interface will handle this
	// go func() {
	// 	ticker := time.NewTicker(5 * time.Minute)
	// 	defer ticker.Stop()
	// 	for range ticker.C {
	// 		if wallesterRepo := repository.NewWallesterRepository(); wallesterRepo != nil {
	// 			if err := wallesterRepo.SyncAllCardsBalances(); err != nil {
	// 				log.Printf("Error syncing all card balances: %v", err)
	// 			}
	// 		}
	// 	}
	// }()

	// 2. Инициализация маршрутизатора (Router)
	router := mux.NewRouter()

	// Health check endpoint для Docker и Render
	router.HandleFunc("/health", handler.HealthCheckHandler).Methods("GET")

	// Настройка публичных маршрутов (Auth)
	router.HandleFunc("/api/v1/auth/register", handler.RegisterHandler).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", handler.LoginHandler).Methods("POST")
	router.HandleFunc("/api/v1/auth/verify-email", handler.VerifyEmailHandler).Methods("GET")
	router.HandleFunc("/api/v1/auth/resend-verification", handler.ResendVerificationHandler).Methods("POST")
	// Rate limiter: 5 запросов сброса пароля за 15 минут с одного IP
	resetLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
	router.HandleFunc("/api/v1/auth/reset-password-request", resetLimiter.Limit(handler.ResetPasswordRequestHandler)).Methods("POST")
	router.HandleFunc("/api/v1/auth/reset-password", handler.ResetPasswordHandler).Methods("POST")
	router.HandleFunc("/api/v1/auth/refresh-token", handler.RefreshTokenHandler).Methods("POST")
	router.HandleFunc("/api/v1/auth/2fa/verify", handler.LoginVerify2FAHandler).Methods("POST")

	// Webhooks (public)
	router.HandleFunc("/api/v1/webhooks/wallester", handler.WallesterWebhookHandler).Methods("POST")
	router.HandleFunc("/api/v1/webhooks/external-topup", handler.ExternalTopUpWebhookHandler).Methods("POST")
	router.HandleFunc("/api/v1/webhooks/sms-receiver", handler.SMSReceiverWebhookHandler).Methods("POST")

	// Telegram Bot Webhook (публичный — Telegram вызывает напрямую)
	router.HandleFunc("/api/v1/telegram/webhook", handler.TelegramWebhookHandler).Methods("POST")

	// Daily report (secret-key protected, for cron/internal use)
	router.HandleFunc("/api/v1/admin/send-daily-report", handler.SendDailyReportHandler).Methods("GET")

	// Notification diagnostic (secret-key protected, for direct testing)
	router.HandleFunc("/api/v1/diag/test-notify", handler.DiagTestNotifyHandler).Methods("GET")

	// Staff PIN verification (JWT-protected, but NOT behind AdminOnly — handler checks is_admin itself)
	router.Handle("/api/v1/verify-staff-pin", middleware.JWTAuthMiddleware(http.HandlerFunc(handler.VerifyStaffPINHandler))).Methods("POST")

	// Public SBP status check
	router.HandleFunc("/api/v1/sbp-status", handler.GetSBPStatusHandler).Methods("GET")

	// Public VPN subscription endpoint (called by v2rayNG / Happ Proxy apps)
	router.HandleFunc("/api/v1/sub/{ref}", handler.VPNSubscriptionHandler).Methods("GET")

	// --- НАСТРОЙКА ЗАЩИЩЕННЫХ МАРШРУТОВ (Protected Routes) ---
	// Создаем Subrouter с префиксом /api/v1/user
	protectedRouter := router.PathPrefix("/api/v1/user").Subrouter()

	// Применяем Middleware ко всем маршрутам, зарегистрированным в protectedRouter
	protectedRouter.Use(middleware.JWTAuthMiddleware)

	// Регистрируем защищенные маршруты
	protectedRouter.HandleFunc("/me", handler.GetMeHandler).Methods("GET")
	protectedRouter.HandleFunc("/grade", handler.GetUserGradeHandler).Methods("GET")
	protectedRouter.HandleFunc("/deposit", handler.ProcessDepositHandler).Methods("POST")
	// Карты и Кошелёк — только для пользователей с подтверждённым email
	verifiedCards := protectedRouter.PathPrefix("/cards").Subrouter()
	verifiedCards.Use(middleware.RequireVerifiedEmail)
	verifiedCards.HandleFunc("", handler.GetUserCardsHandler).Methods("GET")
	verifiedCards.HandleFunc("/issue", handler.MassIssueCardsHandler).Methods("POST")
	verifiedCards.HandleFunc("/{id}/status", handler.PatchCardStatusHandler).Methods("PATCH")
	verifiedCards.HandleFunc("/{id}/auto-replenishment", handler.SetCardAutoReplenishmentHandler).Methods("POST")
	verifiedCards.HandleFunc("/{id}/auto-replenishment", handler.UnsetCardAutoReplenishmentHandler).Methods("DELETE")
	verifiedCards.HandleFunc("/{id}/details", handler.GetCardDetailsHandler).Methods("GET")
	verifiedCards.HandleFunc("/{id}/auto-pay", handler.ToggleAutoPayHandler).Methods("PATCH")
	verifiedCards.HandleFunc("/{id}/subscriptions", handler.CardSubscriptionsHandler).Methods("GET")
	verifiedCards.HandleFunc("/{id}/subscriptions/{subId}", handler.ToggleSubscriptionHandler).Methods("PATCH")
	verifiedCards.HandleFunc("/{id}/freeze-all-subscriptions", handler.FreezeAllSubscriptionsHandler).Methods("POST")
	verifiedCards.HandleFunc("/{id}/sync-balance", handler.SyncCardBalanceHandler).Methods("POST")
	verifiedCards.HandleFunc("/{id}/spending-limit", handler.SetSpendingLimitHandler).Methods("PATCH")

	verifiedWallet := protectedRouter.PathPrefix("/wallet").Subrouter()
	verifiedWallet.Use(middleware.RequireVerifiedEmail)
	verifiedWallet.HandleFunc("", handler.GetWalletHandler).Methods("GET")
	verifiedWallet.HandleFunc("/topup", handler.TopUpWalletHandler).Methods("POST")
	verifiedWallet.HandleFunc("/auto-topup", handler.SetAutoTopupHandler).Methods("PATCH")
	protectedRouter.HandleFunc("/settings/auto-replenish", handler.SetAutoTopupHandler).Methods("PATCH")
	protectedRouter.HandleFunc("/report", handler.GetUserTransactionReportHandler).Methods("GET")
	protectedRouter.HandleFunc("/transactions", handler.GetUnifiedTransactionsHandler).Methods("GET")
	protectedRouter.HandleFunc("/transactions/export", handler.ExportTransactionsHandler).Methods("GET")
	protectedRouter.HandleFunc("/dashboard-stats", handler.GetDashboardStatsHandler).Methods("GET")
	protectedRouter.HandleFunc("/api-key", handler.CreateAPIKeyHandler).Methods("POST")
	protectedRouter.HandleFunc("/upgrade-tier", handler.UpgradeTierHandler).Methods("POST")
	protectedRouter.HandleFunc("/tier-info", handler.GetTierInfoHandler).Methods("GET")

	// Команды (Teams)
	protectedRouter.HandleFunc("/teams", handler.GetUserTeamsHandler).Methods("GET")
	protectedRouter.HandleFunc("/teams", handler.CreateTeamHandler).Methods("POST")
	protectedRouter.HandleFunc("/teams/{id}", handler.GetTeamHandler).Methods("GET")
	protectedRouter.HandleFunc("/teams/{id}/members", handler.InviteTeamMemberHandler).Methods("POST")
	protectedRouter.HandleFunc("/teams/{id}/members/{userId}", handler.RemoveTeamMemberHandler).Methods("DELETE")
	protectedRouter.HandleFunc("/teams/{id}/members/{userId}/role", handler.UpdateTeamMemberRoleHandler).Methods("PATCH")

	// Реферальная программа
	protectedRouter.HandleFunc("/referrals", handler.GetReferralStatsHandler).Methods("GET")
	protectedRouter.HandleFunc("/referrals/info", handler.GetReferralInfoHandler).Methods("GET")

	// Привязка Telegram Chat ID для уведомлений
	protectedRouter.HandleFunc("/settings/telegram", handler.UpdateTelegramChatIDHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/telegram-link", handler.GetTelegramLinkHandler).Methods("GET")
	protectedRouter.HandleFunc("/telegram-status", handler.TelegramStatusHandler).Methods("GET")
	protectedRouter.HandleFunc("/3ds-ws", handler.ThreeDSWebSocketHandler).Methods("GET")

	// Поддержка — отправка тикета
	protectedRouter.HandleFunc("/support", handler.SubmitSupportTicketHandler).Methods("POST")

	// Live Chat
	protectedRouter.HandleFunc("/chat/start", handler.ChatStartHandler).Methods("POST")
	protectedRouter.HandleFunc("/chat/messages/{id}", handler.ChatMessagesHandler).Methods("GET")
	protectedRouter.HandleFunc("/chat/send/{id}", handler.ChatSendHandler).Methods("POST")
	protectedRouter.HandleFunc("/chat/close/{id}", handler.ChatCloseHandler).Methods("POST")
	protectedRouter.HandleFunc("/chat/upload", handler.ChatUploadHandler).Methods("POST")

	// Настройки профиля
	protectedRouter.HandleFunc("/settings/profile", handler.GetSettingsProfileHandler).Methods("GET")
	protectedRouter.HandleFunc("/settings/profile", handler.UpdateProfileHandler).Methods("PATCH")
	protectedRouter.HandleFunc("/settings/change-password", handler.ChangePasswordHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/sessions", handler.GetSessionsHandler).Methods("GET")
	protectedRouter.HandleFunc("/settings/logout-all", handler.LogoutAllSessionsHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/notifications", handler.GetNotificationPrefsHandler).Methods("GET")
	protectedRouter.HandleFunc("/settings/notifications", handler.UpdateNotificationPrefsHandler).Methods("PATCH")
	protectedRouter.HandleFunc("/settings/2fa/setup", handler.Setup2FAHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/2fa/verify", handler.Verify2FAHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/2fa/disable", handler.Disable2FAHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/2fa/unlink", handler.Unlink2FAHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/telegram/unlink", handler.UnlinkTelegramHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/telegram/check-status", handler.CheckTelegramStatusHandler).Methods("GET")
	protectedRouter.HandleFunc("/settings/verify-email-request", handler.RequestEmailVerifyHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/verify-email-confirm", handler.ConfirmEmailVerifyHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/kyc", handler.SubmitKYCHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/kyc", handler.GetKYCHandler).Methods("GET")

	// --- ADMIN ROUTES (JWT + AdminOnly) ---
	adminRouter := router.PathPrefix("/api/v1/admin").Subrouter()
	adminRouter.Use(middleware.JWTAuthMiddleware)
	adminRouter.Use(middleware.AdminOnlyMiddleware)
	adminRouter.HandleFunc("/users", handler.AdminUsersHandler).Methods("GET")
	adminRouter.HandleFunc("/users/{id}/balance", handler.AdminAdjustBalanceHandler).Methods("PATCH")
	adminRouter.HandleFunc("/users/{id}/role", handler.AdminToggleRoleHandler).Methods("PATCH")
	adminRouter.HandleFunc("/users/{id}/status", handler.AdminSetUserStatusHandler).Methods("PATCH")
	adminRouter.HandleFunc("/stats", handler.AdminStatsHandler).Methods("GET")
	adminRouter.HandleFunc("/dashboard", handler.AdminDashboardStatsHandler).Methods("GET")
	adminRouter.HandleFunc("/rates", handler.AdminGetExchangeRatesHandler).Methods("GET")
	adminRouter.HandleFunc("/rates/{id}/markup", handler.AdminUpdateMarkupHandler).Methods("PATCH")
	adminRouter.HandleFunc("/report", handler.GetAdminTransactionReportHandler).Methods("GET")
	adminRouter.HandleFunc("/users/search", handler.AdminSearchUsersHandler).Methods("GET")
	adminRouter.HandleFunc("/users/{id}/grade", handler.AdminUpdateUserGradeHandler).Methods("PATCH")
	adminRouter.HandleFunc("/commissions", handler.AdminGetCommissionConfigHandler).Methods("GET")
	adminRouter.HandleFunc("/commissions/{id}", handler.AdminUpdateCommissionConfigHandler).Methods("PATCH")
	adminRouter.HandleFunc("/tickets", handler.AdminGetSupportTicketsHandler).Methods("GET")
	adminRouter.HandleFunc("/tickets/{id}", handler.AdminUpdateTicketStatusHandler).Methods("PATCH")
	adminRouter.HandleFunc("/users/{id}/full-details", handler.AdminUserFullDetailsHandler).Methods("GET")
	adminRouter.HandleFunc("/users/{id}/emergency-freeze", handler.AdminEmergencyFreezeHandler).Methods("POST")
	adminRouter.HandleFunc("/users/{id}/toggle-block", handler.AdminToggleBlockHandler).Methods("POST")
	adminRouter.HandleFunc("/users/{id}/reset-2fa", handler.AdminReset2FAHandler).Methods("POST")
	adminRouter.HandleFunc("/2fa-status", handler.AdminGet2FAStatusHandler).Methods("GET")
	adminRouter.HandleFunc("/chats", handler.AdminGetChatsHandler).Methods("GET")
	adminRouter.HandleFunc("/chats/{id}/messages", handler.AdminGetChatMessagesHandler).Methods("GET")
	adminRouter.HandleFunc("/translations", handler.AdminGetTranslationsHandler).Methods("GET")
	adminRouter.HandleFunc("/translations", handler.AdminUpsertTranslationHandler).Methods("PUT")
	adminRouter.HandleFunc("/translations/{id}", handler.AdminDeleteTranslationHandler).Methods("DELETE")
	adminRouter.HandleFunc("/logs", handler.AdminGetLogsHandler).Methods("GET")
	adminRouter.HandleFunc("/test-notify", handler.AdminTestNotifyHandler).Methods("GET")
	adminRouter.HandleFunc("/system-settings", handler.GetSystemSettingsHandler).Methods("GET")
	adminRouter.HandleFunc("/system-settings/{key}", handler.UpdateSystemSettingHandler).Methods("PATCH")
	adminRouter.HandleFunc("/infra/balance", handler.GetAezaBalanceHandler).Methods("GET")
	adminRouter.HandleFunc("/infra/balance/check", handler.CheckAezaBalanceHandler).Methods("POST")
	adminRouter.HandleFunc("/infra/server-info", handler.GetAezaServerInfoHandler).Methods("GET")
	adminRouter.HandleFunc("/infra/active-keys", handler.GetActiveVPNKeysHandler).Methods("GET")
	adminRouter.HandleFunc("/infra/vpn-server-status", handler.AdminVPNServerStatusHandler).Methods("GET")
	adminRouter.HandleFunc("/infra/vpn-active-clients", handler.AdminVPNActiveClientsHandler).Methods("GET")
	adminRouter.HandleFunc("/vpn/client/{email}", handler.AdminDeleteVPNClientHandler).Methods("DELETE")
	adminRouter.HandleFunc("/vpn/client/{email}", handler.AdminEditVPNClientHandler).Methods("PATCH")
	// --------------------------------------------------------

	// CORS: dynamic origins from ALLOWED_ORIGINS env var (comma-separated)
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		for _, o := range strings.Split(allowedOrigins, ",") {
			origins = append(origins, strings.TrimSpace(o))
		}
	} else {
		origins = []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"https://xplr-web-dz00cbt2u-djalbens-projects.vercel.app",
			"https://xplr-web.vercel.app",
		}
	}
	log.Printf("CORS allowed origins: %v", origins)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(router)

	// 3. Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s...", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
