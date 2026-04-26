package domain

import (
	"time"

	"gitlab.com/libs-artifex/wrapper/v2"
)

// ErrRateLimited — превышен лимит попыток, надо подождать.
var ErrRateLimited = wrapper.Wrapf(ErrInvalidInput, "too many attempts")

type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	return "too many attempts"
}
