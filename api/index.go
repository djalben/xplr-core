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

	// 6. Auto-migrations (idempotent)
	migrations := []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='cards' AND column_name='category') THEN ALTER TABLE cards ADD COLUMN category VARCHAR(50) DEFAULT 'arbitrage'; END IF; END $$`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			log.Printf("Warning: migration failed: %v", err)
		}
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

	// Wallester webhook (public)
	r.HandleFunc("/api/v1/webhooks/wallester", handlers.WallesterWebhookHandler).Methods("POST")

	// Public card types endpoint
	r.HandleFunc("/api/v1/cards/types", handlers.GetCardTypesHandler).Methods("GET")

	// Protected routes under /api/v1/user
	protected := r.PathPrefix("/api/v1/user").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("/me", handlers.GetMeHandler).Methods("GET")
	protected.HandleFunc("/profile", handlers.GetUserProfileHandler).Methods("GET")
	protected.HandleFunc("/grade", handlers.GetUserGradeHandler).Methods("GET")
	protected.HandleFunc("/deposit", handlers.ProcessDepositHandler).Methods("POST")
	protected.HandleFunc("/topup", handlers.TopUpBalanceHandler).Methods("POST")
	protected.HandleFunc("/cards", handlers.GetUserCardsHandler).Methods("GET")
	protected.HandleFunc("/cards/issue", handlers.MassIssueCardsHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}/status", handlers.PatchCardStatusHandler).Methods("PATCH")
	protected.HandleFunc("/cards/{id}/auto-replenishment", handlers.SetCardAutoReplenishmentHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}/auto-replenishment", handlers.UnsetCardAutoReplenishmentHandler).Methods("DELETE")
	protected.HandleFunc("/cards/{id}/details", handlers.GetCardDetailsHandler).Methods("GET")
	protected.HandleFunc("/cards/{id}/mock-details", handlers.MockCardDetailsHandler).Methods("GET")
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
