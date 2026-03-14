package utils

import (
	"log"

	"gitlab.com/libs-artifex/wrapper/v2"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword хеширует пароль перед сохранением.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)

		return "", wrapper.Wrap(err)
	}

	return string(hashedPassword), nil
}

// CheckPasswordHash сравнивает предоставленный пароль с хешем.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}
