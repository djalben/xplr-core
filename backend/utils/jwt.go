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

// GenerateJWT создает JWT для данного ID пользователя
func GenerateJWT(userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
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
