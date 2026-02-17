package handler

import (
	"context"
	"database/sql"
	"fmt"
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

	// 3. Telegram
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		telegram.SetBotToken(token)
	}

	// 4. Wallester
	handlers.InitWallesterRepository()

	// 5. Start auto-replenishment (runs as goroutine inside the invocation)
	go core.StartAutoReplenishmentWorker()

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

	// Temporary migration endpoint — remove after schema is applied
	r.HandleFunc("/api/migrate", migrateHandler).Methods("POST")

	// Public auth routes
	r.HandleFunc("/api/v1/auth/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/v1/auth/login", handlers.LoginHandler).Methods("POST")

	// Wallester webhook (public)
	r.HandleFunc("/api/v1/webhooks/wallester", handlers.WallesterWebhookHandler).Methods("POST")

	// Protected routes under /api/v1/user
	protected := r.PathPrefix("/api/v1/user").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("/me", handlers.GetMeHandler).Methods("GET")
	protected.HandleFunc("/grade", handlers.GetUserGradeHandler).Methods("GET")
	protected.HandleFunc("/deposit", handlers.ProcessDepositHandler).Methods("POST")
	protected.HandleFunc("/cards", handlers.GetUserCardsHandler).Methods("GET")
	protected.HandleFunc("/cards/issue", handlers.MassIssueCardsHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}/status", handlers.PatchCardStatusHandler).Methods("PATCH")
	protected.HandleFunc("/cards/{id}/auto-replenishment", handlers.SetCardAutoReplenishmentHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}/auto-replenishment", handlers.UnsetCardAutoReplenishmentHandler).Methods("DELETE")
	protected.HandleFunc("/cards/{id}/details", handlers.GetCardDetailsHandler).Methods("GET")
	protected.HandleFunc("/cards/{id}/sync-balance", handlers.SyncCardBalanceHandler).Methods("POST")
	protected.HandleFunc("/report", handlers.GetUserTransactionReportHandler).Methods("GET")
	protected.HandleFunc("/api-key", handlers.CreateAPIKeyHandler).Methods("POST")

	// Teams
	protected.HandleFunc("/teams", handlers.GetUserTeamsHandler).Methods("GET")
	protected.HandleFunc("/teams", handlers.CreateTeamHandler).Methods("POST")
	protected.HandleFunc("/teams/{id}", handlers.GetTeamHandler).Methods("GET")
	protected.HandleFunc("/teams/{id}/members", handlers.InviteTeamMemberHandler).Methods("POST")
	protected.HandleFunc("/teams/{id}/members/{userId}", handlers.RemoveTeamMemberHandler).Methods("DELETE")
	protected.HandleFunc("/teams/{id}/members/{userId}/role", handlers.UpdateTeamMemberRoleHandler).Methods("PATCH")

	// Referrals
	protected.HandleFunc("/referrals", handlers.GetReferralStatsHandler).Methods("GET")

	// Settings
	protected.HandleFunc("/settings/telegram", handlers.UpdateTelegramChatIDHandler).Methods("POST")

	return r
}

func migrateHandler(w http.ResponseWriter, r *http.Request) {
	if repository.GlobalDB == nil {
		http.Error(w, "DB not initialized", http.StatusServiceUnavailable)
		return
	}

	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

		// Drop all tables to reset UUID-based users to SERIAL
		`DROP TABLE IF EXISTS referrals CASCADE`,
		`DROP TABLE IF EXISTS user_grades CASCADE`,
		`DROP TABLE IF EXISTS team_members CASCADE`,
		`DROP TABLE IF EXISTS teams CASCADE`,
		`DROP TABLE IF EXISTS api_keys CASCADE`,
		`DROP TABLE IF EXISTS transactions CASCADE`,
		`DROP TABLE IF EXISTS cards CASCADE`,
		`DROP TABLE IF EXISTS users CASCADE`,

		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			balance NUMERIC(20, 4) DEFAULT 0.0000 NOT NULL,
			balance_rub NUMERIC(20, 4) DEFAULT 0.0000 NOT NULL,
			kyc_status VARCHAR(50) DEFAULT 'pending',
			active_mode VARCHAR(50) DEFAULT 'personal',
			status VARCHAR(50) DEFAULT 'ACTIVE',
			telegram_chat_id BIGINT DEFAULT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='balance') THEN
				ALTER TABLE users ADD COLUMN balance NUMERIC(20,4) DEFAULT 0.0000 NOT NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='balance_rub') THEN
				ALTER TABLE users ADD COLUMN balance_rub NUMERIC(20,4) DEFAULT 0.0000 NOT NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='kyc_status') THEN
				ALTER TABLE users ADD COLUMN kyc_status VARCHAR(50) DEFAULT 'pending';
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='active_mode') THEN
				ALTER TABLE users ADD COLUMN active_mode VARCHAR(50) DEFAULT 'personal';
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='status') THEN
				ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'ACTIVE';
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='telegram_chat_id') THEN
				ALTER TABLE users ADD COLUMN telegram_chat_id BIGINT DEFAULT NULL;
			END IF;
		END $$`,

		`CREATE TABLE IF NOT EXISTS cards (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			provider_card_id VARCHAR(100) NOT NULL,
			bin VARCHAR(6) NOT NULL DEFAULT '424242',
			last_4_digits VARCHAR(4) NOT NULL,
			card_status VARCHAR(50) DEFAULT 'ACTIVE',
			nickname VARCHAR(100),
			daily_spend_limit NUMERIC(20,4) DEFAULT 1000.0000,
			failed_auth_count INTEGER DEFAULT 0,
			card_type VARCHAR(20) DEFAULT 'VISA',
			auto_replenish_enabled BOOLEAN DEFAULT FALSE,
			auto_replenish_threshold NUMERIC(20,4) DEFAULT 0.0000,
			auto_replenish_amount NUMERIC(20,4) DEFAULT 0.0000,
			card_balance NUMERIC(20,4) DEFAULT 0.0000,
			service_slug VARCHAR(50) DEFAULT 'arbitrage',
			team_id INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			card_id INTEGER REFERENCES cards(id),
			amount NUMERIC(20,4) NOT NULL,
			fee NUMERIC(20,4) DEFAULT 0.0000,
			transaction_type VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL,
			details TEXT,
			provider_tx_id VARCHAR(255),
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS api_keys (
			id SERIAL PRIMARY KEY,
			user_id INTEGER,
			api_key UUID UNIQUE DEFAULT uuid_generate_v4(),
			permissions VARCHAR(50) DEFAULT 'READ_ONLY',
			description TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS teams (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			owner_id INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS team_members (
			id SERIAL PRIMARY KEY,
			team_id INTEGER,
			user_id INTEGER,
			role VARCHAR(50) DEFAULT 'member',
			invited_by INTEGER,
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(team_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS user_grades (
			id SERIAL PRIMARY KEY,
			user_id INTEGER UNIQUE,
			grade VARCHAR(50) DEFAULT 'STANDARD',
			total_spent NUMERIC(20,4) DEFAULT 0.0000,
			fee_percent NUMERIC(5,2) DEFAULT 6.70,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS referrals (
			id SERIAL PRIMARY KEY,
			referrer_id INTEGER,
			referred_id INTEGER,
			referral_code VARCHAR(50) UNIQUE NOT NULL,
			status VARCHAR(50) DEFAULT 'PENDING',
			commission_earned NUMERIC(20,4) DEFAULT 0.0000,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
	}

	results := make([]string, 0)
	for i, m := range migrations {
		_, err := repository.GlobalDB.Exec(m)
		if err != nil {
			results = append(results, fmt.Sprintf("migration %d: ERROR: %v", i+1, err))
		} else {
			results = append(results, fmt.Sprintf("migration %d: OK", i+1))
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	for _, r := range results {
		w.Write([]byte(r + "\n"))
	}
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
