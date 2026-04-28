package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/infrastructure/logger"
	"github.com/djalben/xplr-core/backend/internal/transport/http/httpctx"
	"github.com/go-chi/chi/v5/middleware"
)

func WithRequestLogger(l *slog.Logger) func(http.Handler) http.Handler {
	if l == nil {
		l = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ctx := logger.WithLogger(r.Context(), l)
			r = r.WithContext(ctx)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				status := ww.Status()
				if status == 0 {
					status = http.StatusOK
				}

				if status < 400 {
					return
				}

				reqID := middleware.GetReqID(r.Context())
				userID, _ := httpctx.UserID(r.Context())

				l.ErrorContext(
					r.Context(),
					"http request failed",
					"status", status,
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", reqID,
					"user_id", userID.String(),
					"remote_ip", clientIP(r),
					"duration_ms", time.Since(start).Milliseconds(),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func clientIP(r *http.Request) string {
	// middleware.RealIP sets X-Real-IP/X-Forwarded-For into RemoteAddr,
	// but we still normalise host:port.
	addr := strings.TrimSpace(r.RemoteAddr)
	if addr == "" {
		return ""
	}

	host, _, err := net.SplitHostPort(addr)
	if err == nil && host != "" {
		return host
	}

	// already without port
	return addr
}
