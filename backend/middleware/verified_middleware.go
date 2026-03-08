package middleware

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/repository"
)

// RequireVerifiedEmail — middleware, блокирующий доступ к Кошельку и картам
// для пользователей с неподтверждённым email (is_verified = false).
func RequireVerifiedEmail(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok || userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		verified, err := repository.IsUserVerified(userID)
		if err != nil {
			http.Error(w, "Failed to check verification status", http.StatusInternalServerError)
			return
		}

		if !verified {
			http.Error(w, "Email not verified. Please check your inbox and verify your email first.", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
