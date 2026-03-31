package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"gitlab.com/libs-artifex/wrapper/v2"
)

// RandomTokenHex генерирует случайный токен: открытый вид для ссылки и SHA256 в hex для хранения в БД.
func RandomTokenHex(byteLen int) (plain string, hashHex string, err error) {
	b := make([]byte, byteLen)

	_, errRead := rand.Read(b)
	if errRead != nil {
		return "", "", wrapper.Wrap(errRead)
	}

	plain = hex.EncodeToString(b)
	h := sha256.Sum256([]byte(plain))
	hashHex = hex.EncodeToString(h[:])

	return plain, hashHex, nil
}

// HashTokenHex — SHA256(plain) в hex.
func HashTokenHex(plain string) string {
	h := sha256.Sum256([]byte(plain))

	return hex.EncodeToString(h[:])
}
