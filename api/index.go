package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/djalben/xplr-core/backend/vercel"
	"gitlab.com/libs-artifex/wrapper/v2"
)

var (
	handler http.Handler
	initMu  sync.Mutex
	initErr error

	errNilHTTPHandler = errors.New("http handler is nil")
)

func ensureRouterLocked() {
	if handler != nil || initErr != nil {
		return
	}

	func() {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			initErr = wrapper.Wrap(fmt.Errorf("panic during router init: %v\n%s", rec, string(debug.Stack())))
		}()

		h, err := vercel.NewHTTPHandlerFromEnv(context.Background())
		if err != nil {
			initErr = wrapper.Wrap(err)

			return
		}
		if h == nil {
			initErr = wrapper.Wrap(errNilHTTPHandler)

			return
		}

		handler = h
	}()
}

func ensureRouter() {
	initMu.Lock()
	defer initMu.Unlock()

	ensureRouterLocked()
}

// Handler is the Vercel serverless entry point for /api/*.
func Handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		_ = wrapper.Wrap(fmt.Errorf("panic: %v\n%s", rec, string(debug.Stack())))
		http.Error(w, "panic: "+logPanicString(rec), http.StatusInternalServerError)
	}()

	initMu.Lock()
	ensureRouterLocked()
	h := handler
	err := initErr
	initMu.Unlock()

	if err != nil {
		http.Error(w, "init error: "+err.Error(), http.StatusInternalServerError)

		return
	}
	if h == nil {
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

	h.ServeHTTP(w, r)
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
