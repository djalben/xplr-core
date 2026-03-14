package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/internal/app"
	walletApi "github.com/djalben/xplr-core/internal/transport/http/handler/v1/wallet"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Server struct {
	router    *chi.Mux
	container *app.Container
	addr      string
}

// NewServer — создаёт сервер.
func NewServer(container *app.Container, host string, port int) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		container: container,
		addr:      host + ":" + strconv.Itoa(port), // ← исправлено, gosec больше не ругается
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// Start — запускает HTTP-сервер с таймаутами.
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	err := srv.ListenAndServe()
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

	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))
}

// setupRoutes — регистрирует все маршруты.
func (s *Server) setupRoutes() {
	s.router.Route("/v1", func(r chi.Router) {
		walletHandler := walletApi.NewHandler(s.container.WalletUseCase)
		walletHandler.Register(r)
	})
}
