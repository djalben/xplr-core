package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = func() []byte {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret)
	}
	return []byte("my_super_secret_jwt_key")
}()

// GenerateJWT создает JWT с полной информацией о пользователе (id, role, is_admin, token_version).
func GenerateJWT(userID int, isAdmin bool, role string, tokenVersion ...int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id":  userID,
		"is_admin": isAdmin,
		"role":     role,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}
	if len(tokenVersion) > 0 {
		claims["tv"] = tokenVersion[0]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		log.Printf("Error creating JWT: %v", err)
		return "", err
	}
	return tokenString, nil
}

// GenerateHalfAuthJWT creates a short-lived (5min) JWT with half_auth:true.
// Used when user has correct password but still needs 2FA verification.
func GenerateHalfAuthJWT(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"half_auth": true,
		"exp":       time.Now().Add(5 * time.Minute).Unix(),
		"iat":       time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("Error creating half-auth JWT: %v", err)
		return "", err
	}
	return tokenString, nil
}

// ParseHalfAuthJWT validates a half-auth token and returns user_id.
func ParseHalfAuthJWT(tokenString string) (int, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !parsedToken.Valid {
		return 0, fmt.Errorf("invalid half-auth token")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims")
	}
	halfAuth, _ := claims["half_auth"].(bool)
	if !halfAuth {
		return 0, fmt.Errorf("not a half-auth token")
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid user_id in token")
	}
	return int(userIDFloat), nil
}

// GetJWTSecret возвращает секретный ключ для проверки токена
func GetJWTSecret() []byte {
	return jwtKey
}
