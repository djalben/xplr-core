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

	handler.ServeHTTP(w, r)
}

