package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimitEntry struct {
	count    int
	resetAt  time.Time
}

// RateLimiter — простой in-memory rate limiter по IP.
// Ограничивает количество запросов за указанный период.
type RateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*rateLimitEntry
	maxHits  int
	window   time.Duration
}

// NewRateLimiter создаёт rate limiter: maxHits запросов за window.
func NewRateLimiter(maxHits int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		maxHits: maxHits,
		window:  window,
	}
	// Фоновая очистка устаревших записей каждые 5 минут
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			rl.mu.Lock()
			now := time.Now()
			for k, v := range rl.entries {
				if now.After(v.resetAt) {
					delete(rl.entries, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

// Limit — middleware, оборачивает http.HandlerFunc.
// При превышении лимита возвращает 429 Too Many Requests.
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		rl.mu.Lock()
		entry, exists := rl.entries[ip]
		now := time.Now()

		if !exists || now.After(entry.resetAt) {
			rl.entries[ip] = &rateLimitEntry{count: 1, resetAt: now.Add(rl.window)}
			rl.mu.Unlock()
			next(w, r)
			return
		}

		entry.count++
		if entry.count > rl.maxHits {
			rl.mu.Unlock()
			http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
			return
		}
		rl.mu.Unlock()
		next(w, r)
	}
}
