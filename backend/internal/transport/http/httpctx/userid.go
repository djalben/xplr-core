package httpctx

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

type key int

const keyUserID key = iota

func WithUserID(ctx context.Context, id domain.UUID) context.Context {
	return context.WithValue(ctx, keyUserID, id)
}

func UserID(ctx context.Context) (domain.UUID, bool) {
	v, ok := ctx.Value(keyUserID).(domain.UUID)

	return v, ok
}
