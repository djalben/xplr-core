package middleware

import (
	"context"
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
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

// AdminOnly — проверка на админа через UserRepository.
func AdminOnly(userRepo ports.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("userID").(domain.UUID)
			if !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)

				return
			}

			user, err := userRepo.GetByID(r.Context(), userID)
			if err != nil || user == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)

				return
			}

			if !user.IsAdmin {
				http.Error(w, "Forbidden", http.StatusForbidden)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
