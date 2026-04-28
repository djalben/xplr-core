package handler

import (
	"context"
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/infrastructure/logger"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// WrapAndWriteError wraps internal error (for logging/trace) but never exposes it to client.
// Returned error is the wrapped version (so caller may log/inspect it).
func WrapAndWriteError(ctx context.Context, w http.ResponseWriter, err error, status int, publicMessage string) error {
	var wrapped error
	if err != nil {
		wrapped = wrapper.Wrap(err)
		logger.FromContext(ctx).ErrorContext(ctx, "request error", "status", status, "error", wrapped)
	}

	http.Error(w, publicMessage, status)

	return wrapped
}

func WriteBadRequest(w http.ResponseWriter, publicMessage string) {
	http.Error(w, publicMessage, http.StatusBadRequest)
}

func WriteInternalServerError(ctx context.Context, w http.ResponseWriter, err error) error {
	return WrapAndWriteError(ctx, w, err, http.StatusInternalServerError, "Ошибка сервера. Попробуйте позже.")
}
