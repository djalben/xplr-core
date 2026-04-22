package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/djalben/xplr-core/backend/vercel"
)

var (
	handler    http.Handler
	routerOnce sync.Once
	initErr    error

	errNilHTTPHandler = errors.New("http handler is nil")
)

func ensureRouter() {
	routerOnce.Do(func() {
		h, err := vercel.NewHTTPHandlerFromEnv(context.Background())
		if err != nil {
			initErr = err
			log.Printf("router init error: %v", err)

			return
		}
		if h == nil {
			initErr = errNilHTTPHandler
			log.Printf("router init error: %v", initErr)

			return
		}

		handler = h
	})
}

// Handler is the Vercel serverless entry point for /api/*.
func Handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		log.Printf("panic: %v\n%s", rec, string(debug.Stack()))
		http.Error(w, "panic: "+logPanicString(rec), http.StatusInternalServerError)
	}()

	ensureRouter()
	if initErr != nil {
		http.Error(w, "init error: "+initErr.Error(), http.StatusInternalServerError)

		return
	}
	if handler == nil {
		http.Error(w, "init error: handler is nil (this should not happen; check router init logs)", http.StatusInternalServerError)

		return
	}

	// Vercel can invoke the function with a path that may or may not include the
	// function mount prefix. Our chi router is mounted under "/api", so we
	// normalize to always include it.
	if r.URL != nil {
		switch {
		case r.URL.Path == "/api":
			r.URL.Path = "/api/"
		case !strings.HasPrefix(r.URL.Path, "/api/"):
			if strings.HasPrefix(r.URL.Path, "/") {
				r.URL.Path = "/api" + r.URL.Path
			} else {
				r.URL.Path = "/api/" + r.URL.Path
			}
		}
	}

	handler.ServeHTTP(w, r)
}

func logPanicString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case error:
		return x.Error()
	default:
		return "unknown"
	}
}
