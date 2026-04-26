package ports

import (
	"context"
	"time"
)

// AuthRateLimiter — анти-брутфорс на auth эндпоинтах.
// Хранилище должно быть shared (Postgres), чтобы работало в serverless.
type AuthRateLimiter interface {
	Allow(ctx context.Context, key string, now time.Time) (allowed bool, retryAfter time.Duration, err error)
	Fail(ctx context.Context, key string, now time.Time) (retryAfter time.Duration, err error)
	Success(ctx context.Context, key string, now time.Time) error
}
