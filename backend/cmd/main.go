package main

import (
	"context"
	"errors" // ← добавлен для errors.Is
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/djalben/xplr-core/internal/app"
	"github.com/djalben/xplr-core/internal/config"
	logHandler "github.com/djalben/xplr-core/internal/infrastructure/logger/handler"
	httpServer "github.com/djalben/xplr-core/internal/transport/http"
	"gitlab.com/libs-artifex/wrapper/v2"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		panic("failed to parse config: " + err.Error())
	}

	// Логгер
	handler := logHandler.Create(cfg.LogPlain, cfg.LogLevel)
	logger := slog.New(handler)

	logger.Info("🚀 Starting XPLR...")

	// Контейнер
	container, err := app.NewContainer(&cfg)
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

	// Запуск HTTP-сервера
	server := httpServer.NewServer(container, cfg.ServerHost, cfg.ServerPort)

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
}
