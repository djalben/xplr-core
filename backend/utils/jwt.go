package utils

import (
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

// GenerateJWT создает JWT с полной информацией о пользователе (id, role, is_admin).
func GenerateJWT(userID int, isAdmin bool, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id":  userID,
		"is_admin": isAdmin,
		"role":     role,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		log.Printf("Error creating JWT: %v", err)
		return "", err
	}
	return tokenString, nil
}

// GetJWTSecret возвращает секретный ключ для проверки токена
func GetJWTSecret() []byte {
	return jwtKey
}
