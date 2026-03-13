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
			log.Printf("[AUTH-CHECK] ❌ No userID in context for %s %s", r.Method, r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := repository.GetUserByID(userID)
		if err != nil {
			log.Printf("[AUTH-CHECK] ❌ User %d: DB error: %v (path=%s)", userID, err, r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("[AUTH-CHECK] User: %s, is_admin: %v, role: %s, path: %s",
			user.Email, user.IsAdmin, user.Role, r.URL.Path)

		if !user.IsAdmin && user.Role != "admin" {
			log.Printf("[AUTH-CHECK] ⛔ DENIED: user %d (%s) is not admin (is_admin=%v, role=%s)",
				userID, user.Email, user.IsAdmin, user.Role)
			http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
			return
		}

		log.Printf("[AUTH-CHECK] ✅ GRANTED: admin access for %s (user_id=%d)", user.Email, userID)
		next.ServeHTTP(w, r)
	})
}
