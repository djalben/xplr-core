package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/app"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	adminApi "github.com/djalben/xplr-core/backend/internal/transport/http/handler/admin"
	authApi "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/auth"
	authMiddleware "github.com/djalben/xplr-core/backend/internal/transport/http/middleware"
	cardApi "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/card"
	userApi "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/user"
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

func NewServer(container *app.Container, host string, port int, jwtSecret []byte, corsAllowedOrigins string) *Server {
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

	s.setupMiddleware(corsAllowedOrigins)
	s.setupRoutes(jwtSecret)

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

func (s *Server) setupMiddleware(corsAllowedOrigins string) {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	s.router.Use(cors.Handler(buildCORSOptions(corsAllowedOrigins)))
}

// buildCORSOptions — Fetch: Access-Control-Allow-Origin=* несовместим с credentials; для dev без явных origin отключаем credentials.
func buildCORSOptions(originsCSV string) cors.Options {
	opts := cors.Options{
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "Accept", "X-Request-ID"},
		MaxAge:         300,
	}

	list := parseCORSOrigins(originsCSV)
	if len(list) == 0 {
		opts.AllowedOrigins = []string{"*"}
		opts.AllowCredentials = false

		return opts
	}

	opts.AllowedOrigins = list
	opts.AllowCredentials = true

	return opts
}

func parseCORSOrigins(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

func (s *Server) setupRoutes(jwtSecret []byte) {
	s.router.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		r.Route("/v1", func(r chi.Router) {
			// Public: auth
			authApi.NewHandler(s.container.AuthUseCase, s.container.WalletUseCase, s.container.UserRepo, jwtSecret).Register(r)

			// Public: rates (курсы валют)
			r.Get("/rates", func(w http.ResponseWriter, r *http.Request) {
				handler.WriteJSON(w, http.StatusOK, map[string]any{
					"usd": 89.45,
					"eur": 97.82,
				})
			})

			// Protected: user (BFF), wallet, card, transaction, ticket
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.Auth(jwtSecret))
				userApi.NewHandler(s.container.UserUseCase, s.container.WalletUseCase, s.container.GradesUseCase, s.container.CardUseCase, s.container.TransactionUseCase, s.container.TicketUseCase).Register(r)
				walletApi.NewHandler(s.container.WalletUseCase).Register(r)
				cardApi.NewHandler(s.container.CardUseCase).Register(r)
				ticketApi.NewHandler(s.container.TicketUseCase).Register(r)
				transactionApi.NewHandler(s.container.TransactionUseCase).Register(r)
			})
		})

		r.Route("/admin", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.Auth(jwtSecret))
				r.Use(authMiddleware.AdminOnly(s.container.UserRepo))
				adminApi.NewHandler(s.container.CardUseCase, s.container.CommissionUseCase,
					s.container.TicketUseCase, s.container.GradesUseCase).Register(r)
			})
		})
	})
}
