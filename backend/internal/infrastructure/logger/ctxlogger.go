package logger

import (
	"context"
	"log/slog"
)

type ctxKey int

const ctxKeyLogger ctxKey = iota

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	if l == nil {
		return ctx
	}

	return context.WithValue(ctx, ctxKeyLogger, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}

	if l, ok := ctx.Value(ctxKeyLogger).(*slog.Logger); ok && l != nil {
		return l
	}

	return slog.Default()
}
