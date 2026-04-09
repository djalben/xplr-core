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
		// Seed default commission values (idempotent) — only STANDARD and GOLD tiers
		`INSERT INTO commission_config (key, value, description) VALUES
			('fee_standard', 6.70, 'Комиссия за выпуск карты — тир Стандарт ($)'),
			('fee_gold', 4.50, 'Комиссия за выпуск карты — тир Gold ($)'),
			('referral_percent', 5.00, 'Процент реферальной комиссии'),
			('card_issue_fee', 2.00, 'Базовая стоимость выпуска карты ($)')
		ON CONFLICT (key) DO NOTHING`,
		// Seed card configs (idempotent)
		`INSERT INTO card_configs (card_type, issue_fee, transaction_fee_percent, max_single_topup, daily_spend_limit, description) VALUES
			('subscriptions', 2.00, 0.50, 500.00, 300.00, 'Карта для подписок и сервисов'),
			('travel', 3.00, 0.75, 1000.00, 500.00, 'Карта для путешествий'),
			('premium', 5.00, 1.00, 2000.00, 1000.00, 'Премиум карта с расширенными лимитами'),
			('universal', 2.50, 0.60, 750.00, 400.00, 'Универсальная карта')
		ON CONFLICT (card_type) DO NOTHING`,
		// News table
		`CREATE TABLE IF NOT EXISTS news (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			image_url TEXT DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Store tables
		`CREATE TABLE IF NOT EXISTS store_categories (
			id SERIAL PRIMARY KEY,
			slug VARCHAR(50) UNIQUE NOT NULL,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			icon TEXT DEFAULT '',
			sort_order INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS store_products (
			id SERIAL PRIMARY KEY,
			category_id INTEGER REFERENCES store_categories(id),
			provider VARCHAR(50) DEFAULT 'demo',
			external_id TEXT DEFAULT '',
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			country TEXT DEFAULT '',
			country_code VARCHAR(10) DEFAULT '',
			price_usd NUMERIC(10,2) NOT NULL,
			data_gb TEXT DEFAULT '',
			validity_days INTEGER DEFAULT 0,
			image_url TEXT DEFAULT '',
			product_type VARCHAR(30) DEFAULT 'digital',
			in_stock BOOLEAN DEFAULT TRUE,
			meta JSONB DEFAULT '{}',
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS store_orders (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			product_name TEXT NOT NULL,
			price_usd NUMERIC(10,2) NOT NULL,
			status VARCHAR(30) DEFAULT 'pending',
			activation_key TEXT DEFAULT '',
			qr_data TEXT DEFAULT '',
			provider_ref TEXT DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Seed system settings (idempotent)
		`INSERT INTO system_settings (setting_key, setting_value, description) VALUES
			('sbp_enabled', 'true', 'Включить/выключить пополнение через СБП'),
			('gold_tier_price', '50.00', 'Цена апгрейда до Gold (USD)'),
			('gold_tier_duration_days', '30', 'Длительность Gold тира (дней)'),
			('fee_standard', '6.70', 'Комиссия за выпуск карты — тир Стандарт ($)'),
			('fee_gold', '4.50', 'Комиссия за выпуск карты — тир Gold ($)')
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

	// 9c2. CRITICAL: Force currency column in cards table (fixes 500 error on /details)
	if _, err := db.Exec(`ALTER TABLE cards ADD COLUMN IF NOT EXISTS currency TEXT DEFAULT 'USD'`); err != nil {
		log.Printf("[CARDS-MIGRATION] currency ALTER TABLE: %v (may already exist, OK)", err)
	} else {
		log.Println("[CARDS-MIGRATION] ✅ currency column ensured via direct ALTER TABLE")
	}

	// 9c3. Force auto_topup_enabled column in internal_balances (persistent toggle state)
	if _, err := db.Exec(`ALTER TABLE internal_balances ADD COLUMN IF NOT EXISTS auto_topup_enabled BOOLEAN DEFAULT FALSE`); err != nil {
		log.Printf("[WALLET-MIGRATION] auto_topup_enabled ALTER TABLE: %v (may already exist, OK)", err)
	} else {
		log.Println("[WALLET-MIGRATION] ✅ auto_topup_enabled column ensured via direct ALTER TABLE")
	}

	// 9c4. Clean up obsolete tier commission keys from commission_config
	if res, err := db.Exec(`DELETE FROM commission_config WHERE key IN ('fee_silver', 'fee_platinum', 'fee_black')`); err != nil {
		log.Printf("[TIER-CLEANUP] ❌ Failed to delete obsolete commission keys: %v", err)
	} else if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("[TIER-CLEANUP] ✅ Deleted %d obsolete commission keys (fee_silver, fee_platinum, fee_black)", n)
	}
	// Update descriptions for remaining commission keys to Russian
	db.Exec(`UPDATE commission_config SET description = 'Комиссия за выпуск карты — тир Стандарт ($)' WHERE key = 'fee_standard'`)
	db.Exec(`UPDATE commission_config SET description = 'Комиссия за выпуск карты — тир Gold ($)' WHERE key = 'fee_gold'`)
	db.Exec(`UPDATE commission_config SET description = 'Базовая стоимость выпуска карты ($)' WHERE key = 'card_issue_fee'`)

	// 9c5. Backfill tier_expires_at for existing Gold users who have NULL expiry
	if res, err := db.Exec(`UPDATE users SET tier_expires_at = created_at + INTERVAL '365 days' WHERE tier = 'gold' AND tier_expires_at IS NULL`); err != nil {
		log.Printf("[TIER-MIGRATION] ❌ Failed to backfill tier_expires_at: %v", err)
	} else if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("[TIER-MIGRATION] ✅ Backfilled tier_expires_at for %d Gold users", n)
	}

	// 9c6. Add news_notifications_enabled column to users
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS news_notifications_enabled BOOLEAN DEFAULT TRUE`); err != nil {
		log.Printf("[NEWS-MIGRATION] news_notifications_enabled ALTER: %v (may already exist, OK)", err)
	} else {
		log.Println("[NEWS-MIGRATION] ✅ news_notifications_enabled column ensured")
	}

	// 9c7. Seed first news article (idempotent)
	if res, err := db.Exec(`INSERT INTO news (title, content, image_url) 
		SELECT 'Гайд: Как оплачивать AppStore картами XPLR (Visa/Mastercard) через регион Сингапур',
		       'Подробная инструкция по оплате AppStore с помощью карт XPLR через регион Сингапур.\n\n1. Откройте настройки Apple ID\n2. Смените регион на Сингапур\n3. Привяжите карту XPLR (Visa/Mastercard)\n4. Совершайте покупки в AppStore\n\nВажно: убедитесь, что на карте достаточно средств перед покупкой.',
		       ''
		WHERE NOT EXISTS (SELECT 1 FROM news LIMIT 1)`); err != nil {
		log.Printf("[NEWS-SEED] ❌ Failed to seed first news: %v", err)
	} else if n, _ := res.RowsAffected(); n > 0 {
		log.Println("[NEWS-SEED] ✅ Seeded first news article")
	}

	// 9c8. Add last_read_news_id column for unread news badge
	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS last_read_news_id INTEGER DEFAULT 0`); err != nil {
		log.Printf("[NEWS-MIGRATION] last_read_news_id ALTER: %v (may already exist, OK)", err)
	} else {
		log.Println("[NEWS-MIGRATION] ✅ last_read_news_id column ensured")
	}

	// 9c9. Seed store categories (idempotent)
	storeCatSeeds := []struct {
		slug, name, desc, icon string
		sort                   int
	}{
		{"esim", "eSIM — Весь мир", "Мобильный интернет в 190+ странах", "globe", 1},
		{"digital", "Цифровые товары", "Игровые ключи, подписки, пополнения", "gamepad", 2},
	}
	for _, sc := range storeCatSeeds {
		db.Exec(`INSERT INTO store_categories (slug, name, description, icon, sort_order) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (slug) DO NOTHING`,
			sc.slug, sc.name, sc.desc, sc.icon, sc.sort)
	}

	// 9c10. Seed demo store products (idempotent via external_id uniqueness check)
	var esimCatID, digitalCatID int
	db.QueryRow(`SELECT id FROM store_categories WHERE slug='esim'`).Scan(&esimCatID)
	db.QueryRow(`SELECT id FROM store_categories WHERE slug='digital'`).Scan(&digitalCatID)

	if esimCatID > 0 {
		esimProducts := []struct {
			extID, name, country, cc, dataGB string
			price                            float64
			days                             int
		}{
			{"esim-tr-5", "Турция 5 ГБ", "Турция", "TR", "5", 4.50, 30},
			{"esim-tr-10", "Турция 10 ГБ", "Турция", "TR", "10", 7.50, 30},
			{"esim-th-5", "Таиланд 5 ГБ", "Таиланд", "TH", "5", 5.00, 15},
			{"esim-th-10", "Таиланд 10 ГБ", "Таиланд", "TH", "10", 8.00, 30},
			{"esim-eu-5", "Европа 5 ГБ", "Европа (30 стран)", "EU", "5", 6.00, 30},
			{"esim-eu-10", "Европа 10 ГБ", "Европа (30 стран)", "EU", "10", 10.00, 30},
			{"esim-us-5", "США 5 ГБ", "США", "US", "5", 5.50, 30},
			{"esim-us-10", "США 10 ГБ", "США", "US", "10", 9.00, 30},
			{"esim-ae-3", "ОАЭ 3 ГБ", "ОАЭ", "AE", "3", 5.00, 15},
			{"esim-ae-10", "ОАЭ 10 ГБ", "ОАЭ", "AE", "10", 12.00, 30},
			{"esim-jp-5", "Япония 5 ГБ", "Япония", "JP", "5", 6.50, 15},
			{"esim-id-5", "Индонезия 5 ГБ", "Индонезия", "ID", "5", 4.00, 30},
			{"esim-global-1", "Глобальный 1 ГБ", "190+ стран", "GLOBAL", "1", 3.50, 30},
			{"esim-global-5", "Глобальный 5 ГБ", "190+ стран", "GLOBAL", "5", 12.50, 30},
		}
		for i, p := range esimProducts {
			db.Exec(`INSERT INTO store_products (category_id, provider, external_id, name, description, country, country_code, price_usd, data_gb, validity_days, product_type, sort_order)
				SELECT $1,'mobimatter',$2,$3,$4,$5,$6,$7,$8,$9,'esim',$10
				WHERE NOT EXISTS (SELECT 1 FROM store_products WHERE external_id=$2)`,
				esimCatID, p.extID, p.name, p.dataGB+" ГБ мобильного интернета", p.country, p.cc, p.price, p.dataGB, p.days, i)
		}
	}
	if digitalCatID > 0 {
		digProducts := []struct {
			extID, name, desc string
			price             float64
		}{
			{"steam-10", "Steam — $10", "Пополнение кошелька Steam на $10", 10.50},
			{"steam-25", "Steam — $25", "Пополнение кошелька Steam на $25", 25.75},
			{"steam-50", "Steam — $50", "Пополнение кошелька Steam на $50", 51.00},
			{"psn-10", "PlayStation — $10", "PSN Card $10", 10.50},
			{"psn-25", "PlayStation — $25", "PSN Card $25", 25.75},
			{"xbox-10", "Xbox — $10", "Xbox Gift Card $10", 10.50},
			{"xbox-25", "Xbox — $25", "Xbox Gift Card $25", 25.75},
			{"nintendo-10", "Nintendo — $10", "Nintendo eShop $10", 10.50},
			{"spotify-1m", "Spotify Premium — 1 мес", "Подписка Spotify Premium 1 месяц", 9.99},
			{"netflix-1m", "Netflix Standard — 1 мес", "Подписка Netflix Standard 1 месяц", 15.49},
		}
		for i, p := range digProducts {
			db.Exec(`INSERT INTO store_products (category_id, provider, external_id, name, description, price_usd, product_type, sort_order)
				SELECT $1,'razer',$2,$3,$4,$5,'digital',$6
				WHERE NOT EXISTS (SELECT 1 FROM store_products WHERE external_id=$2)`,
				digitalCatID, p.extID, p.name, p.desc, p.price, i)
		}
	}
	log.Println("[STORE-MIGRATION] ✅ Store tables + seed data ensured")

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

	// Start Gold expiry notification worker (daily check)
	service.StartGoldExpiryTicker(db)

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
	protected.HandleFunc("/news", handlers.GetNewsHandler).Methods("GET")
	protected.HandleFunc("/news-notifications", handlers.GetNewsNotificationsHandler).Methods("GET")
	protected.HandleFunc("/news-notifications", handlers.UpdateNewsNotificationsHandler).Methods("PATCH")
	protected.HandleFunc("/news/unread-count", handlers.GetUnreadNewsCountHandler).Methods("GET")
	protected.HandleFunc("/news/mark-as-read", handlers.MarkNewsAsReadHandler).Methods("POST")

	// Store
	protected.HandleFunc("/store/catalog", handlers.StoreCatalogHandler).Methods("GET")
	protected.HandleFunc("/store/purchase", handlers.StorePurchaseHandler).Methods("POST")
	protected.HandleFunc("/store/orders", handlers.StoreOrdersHandler).Methods("GET")

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
	admin.HandleFunc("/news", handlers.AdminCreateNewsHandler).Methods("POST")
	admin.HandleFunc("/news/{id}", handlers.AdminDeleteNewsHandler).Methods("DELETE")
	admin.HandleFunc("/upload-image", handlers.AdminUploadImageHandler).Methods("POST")
	admin.HandleFunc("/test-upload", handlers.AdminTestUploadHandler).Methods("GET")

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
