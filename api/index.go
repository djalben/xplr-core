package handler

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/djalben/xplr-core/backend/vercel"
)

var (
	handler    http.Handler
	routerOnce sync.Once
	initErr    error
)

func ensureRouter() {
	routerOnce.Do(func() {
		h, err := vercel.NewHTTPHandlerFromEnv(context.Background())
		if err != nil {
			initErr = err
			log.Printf("router init error: %v", err)
			return
		}

		handler = h
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
