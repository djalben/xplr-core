package handlers

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

// verifyTOTP validates a 6-digit TOTP code against the secret.
// Accepts current time window and ±1 window for clock drift.
func verifyTOTP(secret, code string) bool {
	now := time.Now().Unix() / 30
	for _, offset := range []int64{-1, 0, 1} {
		if generateTOTP(secret, now+offset) == code {
			return true
		}
	}
	return false
}

func generateTOTP(secret string, counter int64) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff
	otp := truncated % uint32(math.Pow10(6))

	return fmt.Sprintf("%06d", otp)
}
