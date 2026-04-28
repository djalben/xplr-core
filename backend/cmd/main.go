package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djalben/xplr-core/backend/internal/app"
	"github.com/djalben/xplr-core/backend/internal/config"
	logHandler "github.com/djalben/xplr-core/backend/internal/infrastructure/logger/handler"
	httpServer "github.com/djalben/xplr-core/backend/internal/transport/http"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		// Тут логгера ещё нет — пишем в stderr и выходим с кодом 1.
		_, _ = os.Stderr.WriteString("failed to parse config: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Логгер
	handler := logHandler.Create(cfg.LogPlain, cfg.LogLevel)
	logger := slog.New(handler)

	logger.Info("🚀 Starting XPLR...")

	// Контейнер
	container, err := app.NewContainer(context.Background(), &cfg)
	if err != nil {
		logger.Error("failed to create container", "error", err)

		os.Exit(1)
	}

	defer func() {
		err := container.Close()
		if err != nil {
			logger.Error("failed to close container", "error", err)
		}
	}()

	// Запуск сервера
	server := httpServer.NewServer(container, cfg.ServerHost, cfg.ServerPort, []byte(cfg.JWTSecret), cfg.CORSAllowedOrigins, logger)

	go func() {
		err := server.Start()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", wrapper.Wrap(err))
		}
	}()

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()
	logger.Info("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		logger.Error("shutdown failed", "error", wrapper.Wrap(err))
	}
}
