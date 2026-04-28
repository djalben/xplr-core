package vercel

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/app"
	"github.com/djalben/xplr-core/backend/internal/config"
	logHandler "github.com/djalben/xplr-core/backend/internal/infrastructure/logger/handler"
	httpServer "github.com/djalben/xplr-core/backend/internal/transport/http"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// NewHTTPHandlerFromEnv строит http.Handler для serverless окружений (Vercel).
// Важно: пакет публичный, но использует internal компоненты внутри модуля backend.
func NewHTTPHandlerFromEnv(ctx context.Context) (http.Handler, error) {
	cfg, err := config.Parse()
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	container, err := app.NewContainer(ctx, &cfg)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	// Logger for serverless runtime.
	handler := logHandler.Create(cfg.LogPlain, cfg.LogLevel)
	logger := slog.New(handler)

	s := httpServer.NewServer(container, cfg.ServerHost, cfg.ServerPort, []byte(cfg.JWTSecret), cfg.CORSAllowedOrigins, logger)

	return s.Handler(), nil
}
