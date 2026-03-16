package utils

import (
	"log"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

var jwtKey = []byte("my_super_secret_jwt_key")

// GenerateJWT создает JWT для данного ID пользователя.
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

		return "", wrapper.Wrap(err)
	}

	return tokenString, nil
}

// GetJWTSecret возвращает секретный ключ для проверки токена.
func GetJWTSecret() []byte {
	return jwtKey
}

func ValidateJWT(tokenStr string) (domain.UUID, error) {
	// TODO: парсинг JWT с jwt.Parse и извлечение userID
	return domain.NewUUID(), nil // заглушка
}
