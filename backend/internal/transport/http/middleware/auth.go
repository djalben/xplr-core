package middleware

import (
	"context"
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/pkg/utils"
)

// Auth — проверка JWT токена.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}

		// Парсим JWT
		userID, err := utils.ValidateJWT(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)

			return
		}

		// Добавляем userID в контекст
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// AdminOnly — проверка на админа.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// userID := handler.GetUserIDFromContext(r)

		// TODO: проверь роль юзера (из userRepo, user.Role == "admin")
		isAdmin := true // заглушка, замени на реальную проверку

		if !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)

			return
		}

		next.ServeHTTP(w, r)
	})
}
