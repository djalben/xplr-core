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
	"github.com/djalben/xplr-core/backend/core"
	"github.com/djalben/xplr-core/backend/handlers"
	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/telegram"
)

// DB - глобальная переменная для подключения к базе данных (будет использоваться только здесь)
var DB *sql.DB

func main() {
	// 0. Загрузка .env файла (если существует)
	if err := godotenv.Load("backend/.env"); err != nil {
		// Попробовать из текущей директории
		_ = godotenv.Load()
	}

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
	handlers.GlobalDB = DB
	// Передача DB в пакет repository (КРИТИЧЕСКИ ВАЖНО!)
	repository.GlobalDB = DB

	log.Println("Successfully connected to the database!")

	// 1.1. Schema Integrity Guard — проверяет и создаёт отсутствующие колонки
	repository.RunSchemaGuard()

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

	// Инициализация Wallester Repository
	handlers.InitWallesterRepository()
	log.Println("Wallester repository initialized")

	// 1.5. Запуск фонового процесса автопополнения карт
	go core.StartAutoReplenishmentWorker()

	// 1.7. Запуск cron-задачи: возврат остатков истёкших карт в Кошелёк
	go core.StartExpiryReclaimWorker()

	// 1.6. Запуск периодической синхронизации балансов карт из Wallester
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Синхронизация каждые 5 минут
		defer ticker.Stop()
		for range ticker.C {
			if wallesterRepo := repository.NewWallesterRepository(); wallesterRepo != nil {
				if err := wallesterRepo.SyncAllCardsBalances(); err != nil {
					log.Printf("Error syncing all card balances: %v", err)
				}
			}
		}
	}()

	// 2. Инициализация маршрутизатора (Router)
	router := mux.NewRouter()

	// Health check endpoint для Docker и Render
	router.HandleFunc("/health", handlers.HealthCheckHandler).Methods("GET")

	// Настройка публичных маршрутов (Auth)
	router.HandleFunc("/api/v1/auth/register", handlers.RegisterHandler).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/api/v1/auth/verify", handlers.VerifyEmailHandler).Methods("GET")
	// Rate limiter: 5 запросов сброса пароля за 15 минут с одного IP
	resetLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
	router.HandleFunc("/api/v1/auth/reset-password-request", resetLimiter.Limit(handlers.ResetPasswordRequestHandler)).Methods("POST")
	router.HandleFunc("/api/v1/auth/reset-password", handlers.ResetPasswordHandler).Methods("POST")
	router.HandleFunc("/api/v1/auth/refresh-token", handlers.RefreshTokenHandler).Methods("POST")

	// Webhook от Wallester (публичный, без middleware)
	router.HandleFunc("/api/v1/webhooks/wallester", handlers.WallesterWebhookHandler).Methods("POST")

	// Webhook от зарубежной организации — подтверждение пополнения (публичный)
	router.HandleFunc("/api/v1/webhooks/external-topup", handlers.ExternalTopUpWebhookHandler).Methods("POST")

	// Telegram Bot Webhook (публичный — Telegram вызывает напрямую)
	router.HandleFunc("/api/v1/telegram/webhook", handlers.TelegramWebhookHandler).Methods("POST")

	// Daily report (secret-key protected, for cron/internal use)
	router.HandleFunc("/api/v1/admin/send-daily-report", handlers.SendDailyReportHandler).Methods("GET")

	// Notification diagnostic (secret-key protected, for direct testing)
	router.HandleFunc("/api/v1/diag/test-notify", handlers.DiagTestNotifyHandler).Methods("GET")

	// Staff PIN verification (JWT-protected, but NOT behind AdminOnly — handler checks is_admin itself)
	router.Handle("/api/v1/verify-staff-pin", middleware.JWTAuthMiddleware(http.HandlerFunc(handlers.VerifyStaffPINHandler))).Methods("POST")

	// --- НАСТРОЙКА ЗАЩИЩЕННЫХ МАРШРУТОВ (Protected Routes) ---
	// Создаем Subrouter с префиксом /api/v1/user
	protectedRouter := router.PathPrefix("/api/v1/user").Subrouter()

	// Применяем Middleware ко всем маршрутам, зарегистрированным в protectedRouter
	protectedRouter.Use(middleware.JWTAuthMiddleware)

	// Регистрируем защищенные маршруты
	protectedRouter.HandleFunc("/me", handlers.GetMeHandler).Methods("GET")
	protectedRouter.HandleFunc("/grade", handlers.GetUserGradeHandler).Methods("GET")
	protectedRouter.HandleFunc("/deposit", handlers.ProcessDepositHandler).Methods("POST")
	// Карты и Кошелёк — только для пользователей с подтверждённым email
	verifiedCards := protectedRouter.PathPrefix("/cards").Subrouter()
	verifiedCards.Use(middleware.RequireVerifiedEmail)
	verifiedCards.HandleFunc("", handlers.GetUserCardsHandler).Methods("GET")
	verifiedCards.HandleFunc("/issue", handlers.MassIssueCardsHandler).Methods("POST")
	verifiedCards.HandleFunc("/{id}/status", handlers.PatchCardStatusHandler).Methods("PATCH")
	verifiedCards.HandleFunc("/{id}/auto-replenishment", handlers.SetCardAutoReplenishmentHandler).Methods("POST")
	verifiedCards.HandleFunc("/{id}/auto-replenishment", handlers.UnsetCardAutoReplenishmentHandler).Methods("DELETE")
	verifiedCards.HandleFunc("/{id}/details", handlers.GetCardDetailsHandler).Methods("GET")
	verifiedCards.HandleFunc("/{id}/sync-balance", handlers.SyncCardBalanceHandler).Methods("POST")
	verifiedCards.HandleFunc("/{id}/spending-limit", handlers.SetSpendingLimitHandler).Methods("PATCH")

	verifiedWallet := protectedRouter.PathPrefix("/wallet").Subrouter()
	verifiedWallet.Use(middleware.RequireVerifiedEmail)
	verifiedWallet.HandleFunc("", handlers.GetWalletHandler).Methods("GET")
	verifiedWallet.HandleFunc("/topup", handlers.TopUpWalletHandler).Methods("POST")
	verifiedWallet.HandleFunc("/auto-topup", handlers.SetAutoTopupHandler).Methods("PATCH")
	protectedRouter.HandleFunc("/settings/auto-replenish", handlers.SetAutoTopupHandler).Methods("PATCH")
	protectedRouter.HandleFunc("/report", handlers.GetUserTransactionReportHandler).Methods("GET")
	protectedRouter.HandleFunc("/transactions", handlers.GetUnifiedTransactionsHandler).Methods("GET")
	protectedRouter.HandleFunc("/transactions/export", handlers.ExportTransactionsHandler).Methods("GET")
	protectedRouter.HandleFunc("/dashboard-stats", handlers.GetDashboardStatsHandler).Methods("GET")
	protectedRouter.HandleFunc("/api-key", handlers.CreateAPIKeyHandler).Methods("POST")

	// Команды (Teams)
	protectedRouter.HandleFunc("/teams", handlers.GetUserTeamsHandler).Methods("GET")
	protectedRouter.HandleFunc("/teams", handlers.CreateTeamHandler).Methods("POST")
	protectedRouter.HandleFunc("/teams/{id}", handlers.GetTeamHandler).Methods("GET")
	protectedRouter.HandleFunc("/teams/{id}/members", handlers.InviteTeamMemberHandler).Methods("POST")
	protectedRouter.HandleFunc("/teams/{id}/members/{userId}", handlers.RemoveTeamMemberHandler).Methods("DELETE")
	protectedRouter.HandleFunc("/teams/{id}/members/{userId}/role", handlers.UpdateTeamMemberRoleHandler).Methods("PATCH")

	// Реферальная программа
	protectedRouter.HandleFunc("/referrals", handlers.GetReferralStatsHandler).Methods("GET")
	protectedRouter.HandleFunc("/referrals/info", handlers.GetReferralInfoHandler).Methods("GET")

	// Привязка Telegram Chat ID для уведомлений
	protectedRouter.HandleFunc("/settings/telegram", handlers.UpdateTelegramChatIDHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/telegram-link", handlers.GetTelegramLinkHandler).Methods("GET")

	// Поддержка — отправка тикета
	protectedRouter.HandleFunc("/support", handlers.SubmitSupportTicketHandler).Methods("POST")

	// Live Chat
	protectedRouter.HandleFunc("/chat/start", handlers.ChatStartHandler).Methods("POST")
	protectedRouter.HandleFunc("/chat/messages/{id}", handlers.ChatMessagesHandler).Methods("GET")
	protectedRouter.HandleFunc("/chat/send/{id}", handlers.ChatSendHandler).Methods("POST")
	protectedRouter.HandleFunc("/chat/close/{id}", handlers.ChatCloseHandler).Methods("POST")

	// Настройки профиля
	protectedRouter.HandleFunc("/settings/profile", handlers.GetSettingsProfileHandler).Methods("GET")
	protectedRouter.HandleFunc("/settings/profile", handlers.UpdateProfileHandler).Methods("PATCH")
	protectedRouter.HandleFunc("/settings/change-password", handlers.ChangePasswordHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/sessions", handlers.GetSessionsHandler).Methods("GET")
	protectedRouter.HandleFunc("/settings/logout-all", handlers.LogoutAllSessionsHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/notifications", handlers.GetNotificationPrefsHandler).Methods("GET")
	protectedRouter.HandleFunc("/settings/notifications", handlers.UpdateNotificationPrefsHandler).Methods("PATCH")
	protectedRouter.HandleFunc("/settings/2fa/setup", handlers.Setup2FAHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/2fa/verify", handlers.Verify2FAHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/2fa/disable", handlers.Disable2FAHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/2fa/unlink", handlers.Unlink2FAHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/telegram/unlink", handlers.UnlinkTelegramHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/telegram/check-status", handlers.CheckTelegramStatusHandler).Methods("GET")
	protectedRouter.HandleFunc("/settings/verify-email-request", handlers.RequestEmailVerifyHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/verify-email-confirm", handlers.ConfirmEmailVerifyHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/kyc", handlers.SubmitKYCHandler).Methods("POST")
	protectedRouter.HandleFunc("/settings/kyc", handlers.GetKYCHandler).Methods("GET")

	// --- ADMIN ROUTES (JWT + AdminOnly) ---
	adminRouter := router.PathPrefix("/api/v1/admin").Subrouter()
	adminRouter.Use(middleware.JWTAuthMiddleware)
	adminRouter.Use(middleware.AdminOnlyMiddleware)
	adminRouter.HandleFunc("/users", handlers.AdminUsersHandler).Methods("GET")
	adminRouter.HandleFunc("/users/{id}/balance", handlers.AdminAdjustBalanceHandler).Methods("PATCH")
	adminRouter.HandleFunc("/users/{id}/role", handlers.AdminToggleRoleHandler).Methods("PATCH")
	adminRouter.HandleFunc("/users/{id}/status", handlers.AdminSetUserStatusHandler).Methods("PATCH")
	adminRouter.HandleFunc("/stats", handlers.AdminStatsHandler).Methods("GET")
	adminRouter.HandleFunc("/dashboard", handlers.AdminDashboardStatsHandler).Methods("GET")
	adminRouter.HandleFunc("/rates", handlers.AdminGetExchangeRatesHandler).Methods("GET")
	adminRouter.HandleFunc("/rates/{id}/markup", handlers.AdminUpdateMarkupHandler).Methods("PATCH")
	adminRouter.HandleFunc("/report", handlers.GetAdminTransactionReportHandler).Methods("GET")
	adminRouter.HandleFunc("/users/search", handlers.AdminSearchUsersHandler).Methods("GET")
	adminRouter.HandleFunc("/users/{id}/grade", handlers.AdminUpdateUserGradeHandler).Methods("PATCH")
	adminRouter.HandleFunc("/commissions", handlers.AdminGetCommissionConfigHandler).Methods("GET")
	adminRouter.HandleFunc("/commissions/{id}", handlers.AdminUpdateCommissionConfigHandler).Methods("PATCH")
	adminRouter.HandleFunc("/tickets", handlers.AdminGetSupportTicketsHandler).Methods("GET")
	adminRouter.HandleFunc("/tickets/{id}", handlers.AdminUpdateTicketStatusHandler).Methods("PATCH")
	adminRouter.HandleFunc("/users/{id}/full-details", handlers.AdminUserFullDetailsHandler).Methods("GET")
	adminRouter.HandleFunc("/users/{id}/emergency-freeze", handlers.AdminEmergencyFreezeHandler).Methods("POST")
	adminRouter.HandleFunc("/users/{id}/toggle-block", handlers.AdminToggleBlockHandler).Methods("POST")
	adminRouter.HandleFunc("/chats", handlers.AdminGetChatsHandler).Methods("GET")
	adminRouter.HandleFunc("/chats/{id}/messages", handlers.AdminGetChatMessagesHandler).Methods("GET")
	adminRouter.HandleFunc("/translations", handlers.AdminGetTranslationsHandler).Methods("GET")
	adminRouter.HandleFunc("/translations", handlers.AdminUpsertTranslationHandler).Methods("PUT")
	adminRouter.HandleFunc("/translations/{id}", handlers.AdminDeleteTranslationHandler).Methods("DELETE")
	adminRouter.HandleFunc("/logs", handlers.AdminGetLogsHandler).Methods("GET")
	adminRouter.HandleFunc("/test-notify", handlers.AdminTestNotifyHandler).Methods("GET")
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
