package middleware

import (
	"context"
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/pkg/utils"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}

		// Парсим JWT
		userID, err := utils.ValidateJWT(tokenStr) // предположим, что ValidateJWT возвращает userID
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

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = handler.GetUserIDFromContext(r)
		// TODO: проверить, что userID — админ (из userRepo.GetByID(userID).Role == "admin")
		isAdmin := true // заглушка

		if !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)

			return
		}

		next.ServeHTTP(w, r)
	})
}
