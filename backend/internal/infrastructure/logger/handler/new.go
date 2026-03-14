package handler

import (
	"log/slog"
	"os"

	"github.com/djalben/xplr-core/internal/infrastructure/logger"
)

// Create - создание хандлера.
func Create(isPlain bool, level string) slog.Handler {
	if isPlain {
		return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logger.GetLevel(level)})
	}

	return New(os.Stdout, &slog.HandlerOptions{Level: logger.GetLevel(level)})
}
