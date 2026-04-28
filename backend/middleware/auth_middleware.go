package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/djalben/xplr-core/backend/pkg/utils"
	"github.com/djalben/xplr-core/backend/repository"
)

// Контекстный ключ для хранения ID пользователя после проверки токена
type ContextKey string

const UserIDKey ContextKey = "userID"

// JWTAuthMiddleware — авторизация по JWT (Bearer) или по заголовку X-API-Key (таблица api_keys).
// Сначала проверяется X-API-Key (для интеграции арбитражных трекеров), при отсутствии — Bearer JWT.
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID int
		ok := false

		// 1. Попытка авторизации по X-API-Key (для трекеров)
		apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
		if apiKey != "" {
			uid, err := repository.GetUserIDByAPIKey(apiKey)
			if err == nil && uid > 0 {
				userID = uid
				ok = true
			}
		}

		// 2. Если не авторизовались по API Key — проверяем JWT
		if !ok {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization required: use Bearer <token> or X-API-Key header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Authorization format must be Bearer <token>", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					log.Printf("Unexpected signing method: %v", token.Header["alg"])
					return nil, jwt.ErrSignatureInvalid
				}
				return utils.GetJWTSecret(), nil
			})

			if err != nil || !token.Valid {
				log.Printf("Token validation failed: %v", err)
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			claims, okClaims := token.Claims.(jwt.MapClaims)
			if !okClaims {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// SECURITY: Reject half_auth tokens — they must NOT access protected endpoints.
			// Half-auth tokens are only valid for POST /auth/2fa/verify (parsed separately there).
			if halfAuth, _ := claims["half_auth"].(bool); halfAuth {
				log.Printf("[AUTH] ❌ Rejected half_auth token (2FA not completed)")
				http.Error(w, "2FA verification required", http.StatusUnauthorized)
				return
			}

			userIDFloat, okClaims := claims["user_id"].(float64)
			if !okClaims {
				http.Error(w, "User ID not found in token claims", http.StatusUnauthorized)
				return
			}

			uid := int(userIDFloat)

			// SECURITY: Check token_version — reject tokens invalidated by logout-all.
			if tvClaim, hasTv := claims["tv"].(float64); hasTv {
				dbVersion, dbErr := repository.GetTokenVersion(uid)
				if dbErr == nil && int(tvClaim) < dbVersion {
					log.Printf("[AUTH] ❌ Stale token_version for user %d (token=%d, db=%d)", uid, int(tvClaim), dbVersion)
					http.Error(w, "Session expired. Please login again.", http.StatusUnauthorized)
					return
				}
			}

			userID = uid
			ok = true
		}

		if !ok || userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user is blocked
		if repository.IsUserBlocked(userID) {
			http.Error(w, "Ваш аккаунт заблокирован. Пожалуйста, обратитесь в поддержку.", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
