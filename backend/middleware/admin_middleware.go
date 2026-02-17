package middleware

import (
	"log"
	"net/http"

	"github.com/djalben/xplr-core/backend/repository"
)

// AdminOnlyMiddleware checks that the authenticated user has is_admin=true.
// Must be used AFTER JWTAuthMiddleware so that UserIDKey is set in context.
func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok || userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := repository.GetUserByID(userID)
		if err != nil {
			log.Printf("AdminOnly: failed to fetch user %d: %v", userID, err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin {
			log.Printf("AdminOnly: user %d (%s) attempted admin access — denied", userID, user.Email)
			http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
