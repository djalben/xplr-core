package middleware

import (
	"context"
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/pkg/utils"
)

// Auth — проверка JWT токена.
func Auth(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := r.Header.Get("Authorization")
			if tokenStr == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)

				return
			}

			userID, err := utils.ValidateJWT(jwtSecret, tokenStr)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)

				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "userID", userID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// AdminOnly — проверка на админа (заглушка до добавления is_admin в users).
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin := true

		if !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)

			return
		}

		next.ServeHTTP(w, r)
	})
}
