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

func NewServer(container *app.Container, host string, port int) *Server {
	r := chi.NewRouter()

	s := &Server{
		srv: &http.Server{
			Addr:         host + ":" + strconv.Itoa(port),
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		container: container,
		router:    r,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) Start() error {
	err := s.srv.ListenAndServe()
	if err != nil {
		return wrapper.Wrap(err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	err := s.srv.Shutdown(ctx)
	if err != nil {
		return wrapper.Wrap(err)
	}
	return nil
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))
}

func (s *Server) setupRoutes() {
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	s.router.Route("/v1", func(r chi.Router) {
		walletApi.NewHandler(s.container.WalletUseCase).Register(r)
		cardApi.NewHandler(s.container.CardUseCase).Register(r)
		ticketApi.NewHandler(s.container.TicketUseCase).Register(r)
		transactionApi.NewHandler(s.container.TransactionUseCase).Register(r)
	})

	s.router.Route("/admin", func(r chi.Router) {
		// adminApi.NewHandler(...) уже добавлен в предыдущем шаге
		adminApi.NewHandler(s.container.CardUseCase, s.container.CommissionUseCase, s.container.TicketUseCase, s.container.GradesUseCase).Register(r)
	})
}
