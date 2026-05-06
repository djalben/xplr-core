package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func ProviderSecret(secret string) func(http.Handler) http.Handler {
	secret = strings.TrimSpace(secret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" {
				http.Error(w, "Provider secret not configured", http.StatusServiceUnavailable)
				return
			}

			got := strings.TrimSpace(r.Header.Get("X-Provider-Secret"))
			if got == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

