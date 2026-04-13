package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/djalben/xplr-core/backend/internal/app"
	"github.com/djalben/xplr-core/backend/internal/config"
	httpServer "github.com/djalben/xplr-core/backend/internal/transport/http"
)

var (
	handler    http.Handler
	routerOnce sync.Once
	initErr    error
)

func ensureRouter() {
	routerOnce.Do(func() {
		cfg, err := config.Parse()
		if err != nil {
			initErr = err
			log.Printf("config parse error: %v", err)
			return
		}

		container, err := app.NewContainer(&cfg)
		if err != nil {
			initErr = err
			log.Printf("container init error: %v", err)
			return
		}

		// Important: in serverless we don't call ListenAndServe; we use the router directly.
		s := httpServer.NewServer(container, cfg.ServerHost, cfg.ServerPort, []byte(cfg.JWTSecret), cfg.CORSAllowedOrigins)
		handler = s.Handler()
	})
}

// Handler is the Vercel serverless entry point.
func Handler(w http.ResponseWriter, r *http.Request) {
	ensureRouter()
	if initErr != nil {
		http.Error(w, "init error: "+initErr.Error(), http.StatusInternalServerError)
		return
	}
	if handler == nil {
		http.Error(w, "init error: handler is nil", http.StatusInternalServerError)
		return
	}

	// Adapt old API prefix (/api/v1/...) to new server prefix (/api/v1/... is under /api/v1 in server.go's /api/v1 routes).
	// The new server mounts routes under /api/v1 (because it uses /api/v1 via /api/v1 in router: /api/v1 is /api/v1).
	// But internal server.go currently expects /api/v1/* under /api/v1 because it uses /api/v1 via /api/v1? Actually it mounts /api/v1 under /api/v1 by routing /api/v1.
	// We already match by using same paths; no rewrite is needed.
	handler.ServeHTTP(w, r)
}
