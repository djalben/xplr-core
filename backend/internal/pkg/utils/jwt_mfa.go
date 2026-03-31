package utils

import (
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gitlab.com/libs-artifex/wrapper/v2"
)

const mfaPendingClaim = "mfa_pending"

// GenerateMFAPendingJWT — короткоживущий токен после успешного пароля при включённом TOTP.
func GenerateMFAPendingJWT(secret []byte, userID domain.UUID, email string) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := jwt.MapClaims{
		"user_id":       userID.String(),
		"email":         email,
		mfaPendingClaim: true,
		"exp":           expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", wrapper.Wrap(err)
	}

	return tokenString, nil
}

// ValidateMFAPendingJWT — извлекает user_id и email из MFA-токена.
func ValidateMFAPendingJWT(secret []byte, tokenStr string) (domain.UUID, string, error) {
	tokenStr = strings.TrimPrefix(tokenStr, bearerPrefix)

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return secret, nil
	})
	if err != nil {
		return domain.UUID{}, "", wrapper.Wrap(err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return domain.UUID{}, "", wrapper.Wrap(jwt.ErrTokenInvalidClaims)
	}

	mp, ok := claims[mfaPendingClaim].(bool)
	if !ok || !mp {
		return domain.UUID{}, "", wrapper.Wrap(jwt.ErrTokenInvalidClaims)
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok || userIDStr == "" {
		return domain.UUID{}, "", wrapper.Wrap(jwt.ErrTokenInvalidClaims)
	}

	email, _ := claims["email"].(string)

	parsed, err := uuid.Parse(userIDStr)
	if err != nil {
		return domain.UUID{}, "", wrapper.Wrap(err)
	}

	return parsed, email, nil
}
