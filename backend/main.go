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
	"github.com/rs/cors"
	_ "github.com/lib/pq"

	// ВАЖНО: Убедитесь, что пути верны
	"github.com/aalabin/xplr/handlers"
	"github.com/aalabin/xplr/middleware"
	"github.com/aalabin/xplr/repository"
	"github.com/aalabin/xplr/core"
	"github.com/aalabin/xplr/telegram"
)

// DB - глобальная переменная для подключения к базе данных (будет использоваться только здесь)
var DB *sql.DB

func main() {
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

	// Telegram bot token (для реальной отправки уведомлений)
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		telegram.SetBotToken(token)
		log.Println("Telegram bot token set: real notifications enabled")
	}

	// Инициализация Wallester Repository
	handlers.InitWallesterRepository()
	log.Println("Wallester repository initialized")

	// 1.5. Запуск фонового процесса автопополнения карт
	go core.StartAutoReplenishmentWorker()

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

	// Webhook от Wallester (публичный, без middleware)
	router.HandleFunc("/api/v1/webhooks/wallester", handlers.WallesterWebhookHandler).Methods("POST")

	// --- НАСТРОЙКА ЗАЩИЩЕННЫХ МАРШРУТОВ (Protected Routes) ---
	// Создаем Subrouter с префиксом /api/v1/user
	protectedRouter := router.PathPrefix("/api/v1/user").Subrouter()

	// Применяем Middleware ко всем маршрутам, зарегистрированным в protectedRouter
	protectedRouter.Use(middleware.JWTAuthMiddleware)

	// Регистрируем защищенные маршруты
	protectedRouter.HandleFunc("/me", handlers.GetMeHandler).Methods("GET")
	protectedRouter.HandleFunc("/grade", handlers.GetUserGradeHandler).Methods("GET")
	protectedRouter.HandleFunc("/deposit", handlers.ProcessDepositHandler).Methods("POST")
	protectedRouter.HandleFunc("/cards", handlers.GetUserCardsHandler).Methods("GET")
	protectedRouter.HandleFunc("/cards/issue", handlers.MassIssueCardsHandler).Methods("POST")
	protectedRouter.HandleFunc("/cards/{id}/status", handlers.PatchCardStatusHandler).Methods("PATCH")
	protectedRouter.HandleFunc("/cards/{id}/auto-replenishment", handlers.SetCardAutoReplenishmentHandler).Methods("POST")
	protectedRouter.HandleFunc("/cards/{id}/auto-replenishment", handlers.UnsetCardAutoReplenishmentHandler).Methods("DELETE")
	protectedRouter.HandleFunc("/cards/{id}/details", handlers.GetCardDetailsHandler).Methods("GET")
	protectedRouter.HandleFunc("/cards/{id}/sync-balance", handlers.SyncCardBalanceHandler).Methods("POST")
	protectedRouter.HandleFunc("/report", handlers.GetUserTransactionReportHandler).Methods("GET")
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

	// Привязка Telegram Chat ID для уведомлений
	protectedRouter.HandleFunc("/settings/telegram", handlers.UpdateTelegramChatIDHandler).Methods("POST")
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