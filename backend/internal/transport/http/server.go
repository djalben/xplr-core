package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/internal/app"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	adminApi "github.com/djalben/xplr-core/backend/internal/transport/http/handler/admin"
	cardApi "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/card"
	ticketApi "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/ticket"
	transactionApi "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/transaction"
	walletApi "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/wallet"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Server struct {
	srv       *http.Server
	container *app.Container
	router    *chi.Mux
}

// NewServer — создаёт сервер.
func NewServer(container *app.Container, host string, port int) *Server {
	r := chi.NewRouter()

	s := &Server{
		srv: &http.Server{
			Addr:    host + ":" + strconv.Itoa(port),
			Handler: r,
		},
		container: container,
		router:    r,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// Start — запускает сервер.
func (s *Server) Start() error {
	err := s.srv.ListenAndServe()
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// Shutdown — gracefully закрывает сервер.
func (s *Server) Shutdown(ctx context.Context) error {
	err := s.srv.Shutdown(ctx)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

// setupMiddleware — настраивает middleware.
func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second)) // добавлен таймаут для запросов

	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))
}

// setupRoutes — регистрирует маршруты.
func (s *Server) setupRoutes() {
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	s.router.Route("/v1", func(r chi.Router) {
		// Wallet
		walletHandler := walletApi.NewHandler(s.container.WalletUseCase)
		walletHandler.Register(r)

		// Card
		cardHandler := cardApi.NewHandler(s.container.CardUseCase)
		cardHandler.Register(r)

		//ticket
		ticketHandler := ticketApi.NewHandler(s.container.TicketUseCase)
		ticketHandler.Register(r)

		//transaction
		transactionHandler := transactionApi.NewHandler(s.container.TransactionUseCase)
		transactionHandler.Register(r)

	})

	s.router.Route("/admin", func(r chi.Router) {
		// Только для админов (потом middleware)
		adminHandler := adminApi.NewHandler(s.container.CardUseCase, s.container.CommissionUseCase, s.container.TicketUseCase)
		adminHandler.Register(r)
	})
}
