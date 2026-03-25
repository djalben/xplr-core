package handler

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"github.com/djalben/xplr-core/backend/core"
	"github.com/djalben/xplr-core/backend/handlers"
	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/djalben/xplr-core/backend/telegram"
)

var (
	router     *mux.Router
	routerOnce sync.Once
	dbReady    bool
	dbMu       sync.Mutex
)

func ensureDB() {
	dbMu.Lock()
	defer dbMu.Unlock()

	if dbReady {
		return
	}

	// 1. Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("ERROR: DATABASE_URL is not set")
		return
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return
	}

	// Connection pool tuning for serverless (short-lived)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	if err = db.PingContext(context.Background()); err != nil {
		log.Printf("Error pinging database: %v", err)
		return
	}

	// 2. Wire DB into packages
	handlers.GlobalDB = db
	repository.GlobalDB = db

	// 3. Telegram — ОБЯЗАТЕЛЬНО
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if tgToken == "" {
		log.Println("🚨🚨🚨 [FATAL] TELEGRAM_BOT_TOKEN is EMPTY — ALL Telegram notifications are BROKEN")
	} else {
		telegram.SetBotToken(tgToken)
		telegram.AdminChatIDsProvider = repository.GetAdminChatIDs
		log.Printf("✅ [INIT] Telegram bot token set (%d chars)", len(tgToken))
	}
	// SMTP — ОБЯЗАТЕЛЬНО
	if os.Getenv("SMTP_HOST") == "" || os.Getenv("SMTP_PORT") == "" {
		log.Println("🚨🚨🚨 [FATAL] SMTP_HOST/SMTP_PORT not set — ALL email notifications are BROKEN")
	} else if os.Getenv("SMTP_USER") == "" || os.Getenv("SMTP_PASS") == "" {
		log.Println("🚨🚨🚨 [FATAL] SMTP_USER/SMTP_PASS not set — email auth will FAIL")
	} else {
		log.Printf("✅ [INIT] SMTP configured: host=%s, port=%s, user=%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"), os.Getenv("SMTP_USER"))
	}

	// 4. Wallester
	handlers.InitWallesterRepository()

	// 5. Start auto-replenishment (runs as goroutine inside the invocation)
	go core.StartAutoReplenishmentWorker()

	// 6. Auto-migrations (idempotent)
	migrations := []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='cards' AND column_name='category') THEN ALTER TABLE cards ADD COLUMN category VARCHAR(50) DEFAULT 'arbitrage'; END IF; END $$`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='cards' AND column_name='spend_limit') THEN ALTER TABLE cards ADD COLUMN spend_limit NUMERIC(20,4) DEFAULT 0; END IF; END $$`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='is_admin') THEN ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE; END IF; END $$`,
		`CREATE TABLE IF NOT EXISTS referral_codes (
			id SERIAL PRIMARY KEY,
			user_id INTEGER UNIQUE NOT NULL,
			code VARCHAR(20) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS exchange_rates (
			id SERIAL PRIMARY KEY,
			currency_from VARCHAR(10) NOT NULL,
			currency_to VARCHAR(10) NOT NULL,
			base_rate NUMERIC(20,4) NOT NULL DEFAULT 0,
			markup_percent NUMERIC(10,2) NOT NULL DEFAULT 0,
			final_rate NUMERIC(20,4) NOT NULL DEFAULT 0,
			updated_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(currency_from, currency_to)
		)`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='balance_arbitrage') THEN ALTER TABLE users ADD COLUMN balance_arbitrage NUMERIC(20,4) DEFAULT 0; END IF; END $$`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='balance_personal') THEN ALTER TABLE users ADD COLUMN balance_personal NUMERIC(20,4) DEFAULT 0; END IF; END $$`,
		`UPDATE users SET balance_arbitrage = COALESCE(balance, 0) WHERE balance_arbitrage = 0 AND balance > 0`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			log.Printf("Warning: migration failed: %v", err)
		}
	}

	// 7. Create tables that schema_guard doesn't cover (tables, not columns)
	tableMigrations := []string{
		// Wallet (internal balances)
		`CREATE TABLE IF NOT EXISTS internal_balances (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE UNIQUE,
			master_balance NUMERIC(20,4) DEFAULT 0.0000 NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Support tickets
		`CREATE TABLE IF NOT EXISTS support_tickets (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			subject VARCHAR(500) NOT NULL,
			status VARCHAR(50) DEFAULT 'open',
			tg_chat_id BIGINT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Admin logs
		`CREATE TABLE IF NOT EXISTS admin_logs (
			id SERIAL PRIMARY KEY,
			admin_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			action TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Commission config
		`CREATE TABLE IF NOT EXISTS commission_config (
			id SERIAL PRIMARY KEY,
			key VARCHAR(100) UNIQUE NOT NULL,
			value NUMERIC(20,4) NOT NULL,
			description TEXT,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Telegram link codes (for /start UUID deep linking)
		`CREATE TABLE IF NOT EXISTS telegram_link_codes (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE UNIQUE,
			code VARCHAR(64) NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`,
		// User sessions
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			ip VARCHAR(50) DEFAULT '',
			device TEXT DEFAULT '',
			location TEXT DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			last_active TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// KYC requests
		`CREATE TABLE IF NOT EXISTS kyc_requests (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			country VARCHAR(10) NOT NULL,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			birth_date VARCHAR(20),
			address TEXT,
			doc_passport VARCHAR(500),
			doc_address VARCHAR(500),
			doc_selfie VARCHAR(500),
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Card configs - flexible fees and limits per card type
		`CREATE TABLE IF NOT EXISTS card_configs (
			id SERIAL PRIMARY KEY,
			card_type VARCHAR(50) UNIQUE NOT NULL,
			issue_fee NUMERIC(10,2) DEFAULT 2.00,
			transaction_fee_percent NUMERIC(5,2) DEFAULT 0.00,
			max_single_topup NUMERIC(20,4) DEFAULT 1000.0000,
			daily_spend_limit NUMERIC(20,4) DEFAULT 500.0000,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// System settings - global toggles
		`CREATE TABLE IF NOT EXISTS system_settings (
			id SERIAL PRIMARY KEY,
			setting_key VARCHAR(100) UNIQUE NOT NULL,
			setting_value TEXT NOT NULL,
			description TEXT,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Seed default commission values (idempotent)
		`INSERT INTO commission_config (key, value, description) VALUES
			('fee_standard', 6.70, 'Комиссия для грейда STANDARD (%)'),
			('fee_silver', 5.50, 'Комиссия для грейда SILVER (%)'),
			('fee_gold', 4.50, 'Комиссия для грейда GOLD (%)'),
			('fee_platinum', 3.50, 'Комиссия для грейда PLATINUM (%)'),
			('fee_black', 2.50, 'Комиссия для грейда BLACK (%)'),
			('referral_percent', 5.00, 'Процент реферальной комиссии'),
			('card_issue_fee', 2.00, 'Стоимость выпуска карты ($)')
		ON CONFLICT (key) DO NOTHING`,
		// Seed card configs (idempotent)
		`INSERT INTO card_configs (card_type, issue_fee, transaction_fee_percent, max_single_topup, daily_spend_limit, description) VALUES
			('subscriptions', 2.00, 0.50, 500.00, 300.00, 'Карта для подписок и сервисов'),
			('travel', 3.00, 0.75, 1000.00, 500.00, 'Карта для путешествий'),
			('premium', 5.00, 1.00, 2000.00, 1000.00, 'Премиум карта с расширенными лимитами'),
			('universal', 2.50, 0.60, 750.00, 400.00, 'Универсальная карта')
		ON CONFLICT (card_type) DO NOTHING`,
		// Seed system settings (idempotent)
		`INSERT INTO system_settings (setting_key, setting_value, description) VALUES
			('sbp_enabled', 'true', 'Включить/выключить пополнение через СБП'),
			('gold_tier_price', '50.00', 'Цена апгрейда до Gold tier (USD)'),
			('gold_tier_duration_days', '30', 'Длительность Gold tier в днях')
		ON CONFLICT (setting_key) DO NOTHING`,
	}
	for _, m := range tableMigrations {
		if _, err := db.Exec(m); err != nil {
			log.Printf("Warning: table migration failed: %v", err)
		}
	}

	// 8. Run SchemaGuard to ensure all required columns exist
	repository.RunSchemaGuard()

	// 9b. Chat tables
	if err := repository.EnsureChatTables(); err != nil {
		log.Printf("Warning: could not ensure chat tables: %v", err)
	}

	// 9b2. Translations table
	if err := repository.EnsureTranslationsTable(); err != nil {
		log.Printf("Warning: could not ensure translations table: %v", err)
	}

	// 9c. HARD migration: force claimed_by column (DO $$ may fail on Vercel)
	if _, err := db.Exec(`ALTER TABLE chat_conversations ADD COLUMN IF NOT EXISTS claimed_by INTEGER DEFAULT 0`); err != nil {
		log.Printf("[CHAT-MIGRATION] claimed_by ALTER TABLE: %v (may already exist, OK)", err)
	} else {
		log.Println("[CHAT-MIGRATION] ✅ claimed_by column ensured via direct ALTER TABLE")
	}

	// 9d. Force admin rights for known admins
	adminEmails := []string{"aalabin5@gmail.com", "vardump@inbox.ru"}
	for _, email := range adminEmails {
		res, err := db.Exec(`UPDATE users SET is_admin = TRUE WHERE email = $1 AND (is_admin IS NULL OR is_admin = FALSE)`, email)
		if err != nil {
			log.Printf("[ADMIN-SETUP] ❌ Failed to set is_admin for %s: %v", email, err)
		} else if n, _ := res.RowsAffected(); n > 0 {
			log.Printf("[ADMIN-SETUP] ✅ is_admin set to TRUE for %s", email)
		}
		// Check telegram_chat_id binding
		var tgChatID int64
		var userID int
		scanErr := db.QueryRow(`SELECT id, COALESCE(telegram_chat_id, 0) FROM users WHERE email = $1`, email).Scan(&userID, &tgChatID)
		if scanErr != nil {
			log.Printf("[ADMIN-SETUP] ⚠️ User %s NOT FOUND in DB: %v", email, scanErr)
		} else if tgChatID == 0 {
			log.Printf("[ADMIN-SETUP] ⚠️ User %s (id=%d) has NO telegram_chat_id — Telegram features will NOT work!", email, userID)
		} else {
			log.Printf("[ADMIN-SETUP] ✅ User %s (id=%d) telegram_chat_id=%d — OK", email, userID, tgChatID)
		}
	}

	// 9. Seed default exchange rates & start fetcher
	repository.SeedDefaultExchangeRates()
	go service.StartExchangeRateFetcher()

	// 10. SMTP diagnostics (log config status, never log passwords)
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpSupportUser := os.Getenv("SMTP_SUPPORT_USER")
	if smtpHost != "" && smtpUser != "" && smtpPass != "" {
		log.Printf("[SMTP] ✅ Configured: host=%s, port=%s, user=%s", smtpHost, smtpPort, smtpUser)
	} else {
		log.Printf("[SMTP] ⚠️  NOT configured! host=%q, port=%q, user=%q, pass_len=%d — emails will FAIL",
			smtpHost, smtpPort, smtpUser, len(smtpPass))
	}
	if smtpSupportUser != "" {
		log.Printf("[SMTP] ✅ Support account: %s", smtpSupportUser)
	} else {
		log.Printf("[SMTP] ℹ️  No SMTP_SUPPORT_USER — support emails will use main SMTP account")
	}

	dbReady = true
	log.Println("Serverless handler initialized successfully")
}

func ensureRouter() {
	routerOnce.Do(func() {
		router = buildRouter()
	})
}

func buildRouter() *mux.Router {
	r := mux.NewRouter()

	// Health
	r.HandleFunc("/api/health", handlers.HealthCheckHandler).Methods("GET")

	// Public auth routes
	r.HandleFunc("/api/v1/auth/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/v1/auth/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/api/v1/auth/verify", handlers.VerifyEmailHandler).Methods("GET")
	r.HandleFunc("/api/v1/auth/reset-password-request", handlers.ResetPasswordRequestHandler).Methods("POST")
	r.HandleFunc("/api/v1/auth/reset-password", handlers.ResetPasswordHandler).Methods("POST")
	r.HandleFunc("/api/v1/auth/refresh-token", handlers.RefreshTokenHandler).Methods("POST")

	// Webhooks (public)
	r.HandleFunc("/api/v1/webhooks/wallester", handlers.WallesterWebhookHandler).Methods("POST")
	r.HandleFunc("/api/v1/webhooks/external-topup", handlers.ExternalTopUpWebhookHandler).Methods("POST")

	// Telegram Bot Webhook (public — Telegram calls directly)
	r.HandleFunc("/api/v1/telegram/webhook", handlers.TelegramWebhookHandler).Methods("POST")

	// Daily report (secret-key protected, for cron/internal use)
	r.HandleFunc("/api/v1/admin/send-daily-report", handlers.SendDailyReportHandler).Methods("GET")

	// Notification diagnostic (secret-key protected)
	r.HandleFunc("/api/v1/diag/test-notify", handlers.DiagTestNotifyHandler).Methods("GET")

	// Public card types endpoint
	r.HandleFunc("/api/v1/cards/types", handlers.GetCardTypesHandler).Methods("GET")

	// Public exchange rates
	r.HandleFunc("/api/v1/rates", handlers.PublicGetExchangeRatesHandler).Methods("GET")

	// Public SBP status check
	r.HandleFunc("/api/v1/sbp-status", handlers.GetSBPStatusHandler).Methods("GET")

	// Staff PIN verification (JWT-protected, NOT behind AdminOnly — handler checks is_admin itself)
	r.Handle("/api/v1/verify-staff-pin", middleware.JWTAuthMiddleware(http.HandlerFunc(handlers.VerifyStaffPINHandler))).Methods("POST")
	log.Println("Registered route: POST /api/v1/verify-staff-pin")

	// Protected routes under /api/v1/user
	protected := r.PathPrefix("/api/v1/user").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("/me", handlers.GetMeHandler).Methods("GET")
	protected.HandleFunc("/profile", handlers.GetUserProfileHandler).Methods("GET")
	protected.HandleFunc("/grade", handlers.GetUserGradeHandler).Methods("GET")
	protected.HandleFunc("/deposit", handlers.ProcessDepositHandler).Methods("POST")
	protected.HandleFunc("/topup", handlers.TopUpBalanceHandler).Methods("POST")
	protected.HandleFunc("/stats", handlers.GetUserStatsHandler).Methods("GET")
	protected.HandleFunc("/cards", handlers.GetUserCardsHandler).Methods("GET")
	protected.HandleFunc("/cards/issue", handlers.MassIssueCardsHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}/status", handlers.PatchCardStatusHandler).Methods("PATCH")
	protected.HandleFunc("/cards/{id}/auto-replenishment", handlers.SetCardAutoReplenishmentHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}/auto-replenishment", handlers.UnsetCardAutoReplenishmentHandler).Methods("DELETE")
	protected.HandleFunc("/cards/{id}/details", handlers.GetCardDetailsHandler).Methods("GET")
	protected.HandleFunc("/cards/{id}/mock-details", handlers.MockCardDetailsHandler).Methods("GET")
	protected.HandleFunc("/cards/{id}/limit", handlers.UpdateCardSpendLimitHandler).Methods("PATCH")
	protected.HandleFunc("/cards/{id}/sync-balance", handlers.SyncCardBalanceHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}/spending-limit", handlers.SetSpendingLimitHandler).Methods("PATCH")
	protected.HandleFunc("/wallet", handlers.GetWalletHandler).Methods("GET")
	protected.HandleFunc("/wallet/topup", handlers.TopUpWalletHandler).Methods("POST")
	protected.HandleFunc("/wallet/transfer-to-card", handlers.TransferWalletToCardHandler).Methods("POST")
	protected.HandleFunc("/wallet/auto-topup", handlers.SetAutoTopupHandler).Methods("PATCH")
	protected.HandleFunc("/report", handlers.GetUserTransactionReportHandler).Methods("GET")
	protected.HandleFunc("/transactions", handlers.GetUnifiedTransactionsHandler).Methods("GET")
	protected.HandleFunc("/transactions/export", handlers.ExportTransactionsHandler).Methods("GET")
	protected.HandleFunc("/dashboard-stats", handlers.GetDashboardStatsHandler).Methods("GET")
	protected.HandleFunc("/settings/auto-replenish", handlers.SetAutoTopupHandler).Methods("PATCH")
	protected.HandleFunc("/api-key", handlers.CreateAPIKeyHandler).Methods("POST")
	protected.HandleFunc("/upgrade-tier", handlers.UpgradeTierHandler).Methods("POST")
	protected.HandleFunc("/tier-info", handlers.GetTierInfoHandler).Methods("GET")
	log.Println("Registered route: GET /api/v1/user/dashboard-stats")

	// Teams
	protected.HandleFunc("/teams", handlers.GetUserTeamsHandler).Methods("GET")
	protected.HandleFunc("/teams", handlers.CreateTeamHandler).Methods("POST")
	protected.HandleFunc("/teams/{id}", handlers.GetTeamHandler).Methods("GET")
	protected.HandleFunc("/teams/{id}/members", handlers.InviteTeamMemberHandler).Methods("POST")
	protected.HandleFunc("/teams/{id}/members/{userId}", handlers.RemoveTeamMemberHandler).Methods("DELETE")
	protected.HandleFunc("/teams/{id}/members/{userId}/role", handlers.UpdateTeamMemberRoleHandler).Methods("PATCH")

	// Referrals
	protected.HandleFunc("/referrals", handlers.GetReferralStatsHandler).Methods("GET")
	protected.HandleFunc("/referrals/info", handlers.GetReferralInfoHandler).Methods("GET")
	protected.HandleFunc("/referrals/list", handlers.GetReferralListHandler).Methods("GET")

	// Settings — Telegram
	protected.HandleFunc("/settings/telegram", handlers.UpdateTelegramChatIDHandler).Methods("POST")
	protected.HandleFunc("/settings/telegram-link", handlers.GetTelegramLinkHandler).Methods("GET")

	// Support
	protected.HandleFunc("/support", handlers.SubmitSupportTicketHandler).Methods("POST")

	// Live Chat
	protected.HandleFunc("/chat/start", handlers.ChatStartHandler).Methods("POST")
	protected.HandleFunc("/chat/messages/{id}", handlers.ChatMessagesHandler).Methods("GET")
	protected.HandleFunc("/chat/send/{id}", handlers.ChatSendHandler).Methods("POST")
	protected.HandleFunc("/chat/close/{id}", handlers.ChatCloseHandler).Methods("POST")

	// Settings — Profile, Password, Sessions, Notifications, 2FA, Email Verify, KYC
	protected.HandleFunc("/settings/profile", handlers.GetSettingsProfileHandler).Methods("GET")
	protected.HandleFunc("/settings/profile", handlers.UpdateProfileHandler).Methods("PATCH")
	protected.HandleFunc("/settings/change-password", handlers.ChangePasswordHandler).Methods("POST")
	protected.HandleFunc("/settings/sessions", handlers.GetSessionsHandler).Methods("GET")
	protected.HandleFunc("/settings/logout-all", handlers.LogoutAllSessionsHandler).Methods("POST")
	protected.HandleFunc("/settings/notifications", handlers.GetNotificationPrefsHandler).Methods("GET")
	protected.HandleFunc("/settings/notifications", handlers.UpdateNotificationPrefsHandler).Methods("PATCH")
	protected.HandleFunc("/settings/2fa/setup", handlers.Setup2FAHandler).Methods("POST")
	protected.HandleFunc("/settings/2fa/verify", handlers.Verify2FAHandler).Methods("POST")
	protected.HandleFunc("/settings/2fa/disable", handlers.Disable2FAHandler).Methods("POST")
	protected.HandleFunc("/settings/2fa/unlink", handlers.Unlink2FAHandler).Methods("POST")
	protected.HandleFunc("/settings/telegram/unlink", handlers.UnlinkTelegramHandler).Methods("POST")
	protected.HandleFunc("/settings/telegram/check-status", handlers.CheckTelegramStatusHandler).Methods("GET")
	protected.HandleFunc("/settings/verify-email-request", handlers.RequestEmailVerifyHandler).Methods("POST")
	protected.HandleFunc("/settings/verify-email-confirm", handlers.ConfirmEmailVerifyHandler).Methods("POST")
	protected.HandleFunc("/settings/kyc", handlers.SubmitKYCHandler).Methods("POST")
	protected.HandleFunc("/settings/kyc", handlers.GetKYCHandler).Methods("GET")

	// Admin routes (JWT + AdminOnly)
	admin := r.PathPrefix("/api/v1/admin").Subrouter()
	admin.Use(middleware.JWTAuthMiddleware)
	admin.Use(middleware.AdminOnlyMiddleware)
	admin.HandleFunc("/stats", handlers.AdminStatsHandler).Methods("GET")
	admin.HandleFunc("/users", handlers.AdminUsersHandler).Methods("GET")
	admin.HandleFunc("/users/{id}/balance", handlers.AdminAdjustBalanceHandler).Methods("PATCH")
	admin.HandleFunc("/users/{id}/role", handlers.AdminToggleRoleHandler).Methods("PATCH")
	admin.HandleFunc("/users/{id}/status", handlers.AdminSetUserStatusHandler).Methods("PATCH")
	admin.HandleFunc("/dashboard", handlers.AdminDashboardStatsHandler).Methods("GET")
	admin.HandleFunc("/rates", handlers.AdminGetExchangeRatesHandler).Methods("GET")
	admin.HandleFunc("/rates/{id}/markup", handlers.AdminUpdateMarkupHandler).Methods("PATCH")
	admin.HandleFunc("/report", handlers.GetAdminTransactionReportHandler).Methods("GET")
	admin.HandleFunc("/users/search", handlers.AdminSearchUsersHandler).Methods("GET")
	admin.HandleFunc("/users/{id}/grade", handlers.AdminUpdateUserGradeHandler).Methods("PATCH")
	admin.HandleFunc("/commissions", handlers.AdminGetCommissionConfigHandler).Methods("GET")
	admin.HandleFunc("/commissions/{id}", handlers.AdminUpdateCommissionConfigHandler).Methods("PATCH")
	admin.HandleFunc("/tickets", handlers.AdminGetSupportTicketsHandler).Methods("GET")
	admin.HandleFunc("/tickets/{id}", handlers.AdminUpdateTicketStatusHandler).Methods("PATCH")
	admin.HandleFunc("/users/{id}/full-details", handlers.AdminUserFullDetailsHandler).Methods("GET")
	admin.HandleFunc("/users/{id}/emergency-freeze", handlers.AdminEmergencyFreezeHandler).Methods("POST")
	admin.HandleFunc("/users/{id}/toggle-block", handlers.AdminToggleBlockHandler).Methods("POST")
	admin.HandleFunc("/chats", handlers.AdminGetChatsHandler).Methods("GET")
	admin.HandleFunc("/chats/{id}/messages", handlers.AdminGetChatMessagesHandler).Methods("GET")
	admin.HandleFunc("/translations", handlers.AdminGetTranslationsHandler).Methods("GET")
	admin.HandleFunc("/translations", handlers.AdminUpsertTranslationHandler).Methods("PUT")
	admin.HandleFunc("/translations/{id}", handlers.AdminDeleteTranslationHandler).Methods("DELETE")
	admin.HandleFunc("/logs", handlers.AdminGetLogsHandler).Methods("GET")
	admin.HandleFunc("/test-notify", handlers.AdminTestNotifyHandler).Methods("GET")
	admin.HandleFunc("/system-settings", handlers.GetSystemSettingsHandler).Methods("GET")
	admin.HandleFunc("/system-settings/{key}", handlers.UpdateSystemSettingHandler).Methods("PATCH")

	log.Println("✅ [ROUTER] All routes registered successfully")
	return r
}

// Handler is the Vercel serverless entry point.
func Handler(w http.ResponseWriter, r *http.Request) {
	ensureRouter()
	ensureDB()

	// CORS headers (same-origin on Vercel, but keep for local dev / preview URLs)
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, X-API-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "300")
	}

	// Handle preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If DB is not ready, return 503 with a clear message (except for health check)
	if !dbReady && r.URL.Path != "/api/health" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"Database not initialized. Check DATABASE_URL environment variable."}`))
		return
	}

	router.ServeHTTP(w, r)
}
