package utils

import (
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gitlab.com/libs-artifex/wrapper/v2"
)

const bearerPrefix = "Bearer "

// GenerateJWT создаёт JWT для данного пользователя.
func GenerateJWT(secret []byte, userID domain.UUID, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"exp":     expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", wrapper.Wrap(err)
	}

	return tokenString, nil
}

// ValidateJWT парсит Bearer токен и возвращает user_id.
func ValidateJWT(secret []byte, tokenStr string) (domain.UUID, error) {
	tokenStr = strings.TrimPrefix(tokenStr, bearerPrefix)
	if tokenStr == "" {
		return domain.UUID{}, wrapper.Wrap(jwt.ErrTokenMalformed)
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return secret, nil
	})
	if err != nil {
		return domain.UUID{}, wrapper.Wrap(err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return domain.UUID{}, wrapper.Wrap(jwt.ErrTokenInvalidClaims)
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok || userIDStr == "" {
		return domain.UUID{}, wrapper.Wrap(jwt.ErrTokenInvalidClaims)
	}

	parsed, err := uuid.Parse(userIDStr)
	if err != nil {
		return domain.UUID{}, wrapper.Wrap(err)
	}

	return parsed, nil
}
