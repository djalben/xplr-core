package handler

import (
	"testing"
)

// ── Test: SMS code extraction from Armenian bank messages ──
func TestExtractSMSCode(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		// Standard code patterns
		{"Код подтверждения: 123456", "123456"},
		{"Your OTP code: 654321", "654321"},
		{"code 789012", "789012"},
		{"PIN: 5678", "5678"},

		// Armenian bank patterns
		{"Ameriabank: Ваш код подтверждения 483921", "483921"},
		{"Ardshinbank OTP: 192837", "192837"},
		{"ACBA 3DS verification code 847261", "847261"},
		{"IDBank payment code: 938471", "938471"},
		{"Evoca: подтверждение 628193", "628193"},

		// 3DS specific
		{"3DS код: 847291", "847291"},
		{"3ds authentication 192384", "192384"},

		// Generic 6-digit code in message
		{"Покупка на сумму 100 USD. Код 847291 для подтверждения.", "847291"},

		// Password pattern
		{"Одноразовый пароль: 928374", "928374"},

		// No code found
		{"Welcome to our service!", ""},
		{"Your balance is 100 USD", ""},
	}

	for _, tt := range tests {
		result := ExtractSMSCode(tt.message)
		if result != tt.expected {
			t.Errorf("ExtractSMSCode(%q) = %q, want %q", tt.message, result, tt.expected)
		}
	}
}

// ── Test: Merchant extraction from SMS ──
func TestExtractMerchantFromSMS(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"оплата в Netflix на сумму 15.99 USD", "Netflix на"},
		{"payment at Amazon.com for 29.99", "Amazon.com for"},
		{"покупка в Google Play", "Google Play"},
		{"Hello world", ""},
	}

	for _, tt := range tests {
		result := extractMerchantFromSMS(tt.message)
		if result != tt.expected {
			t.Errorf("extractMerchantFromSMS(%q) = %q, want %q", tt.message, result, tt.expected)
		}
	}
}

// ── Test: Deliver returns false when no clients ──
func TestThreeDSHub_DeliverNoClients(t *testing.T) {
	delivered := ThreeDSHub.Deliver(99999, SMSCodeMessage{
		Code:         "123456",
		MerchantName: "Test",
	})
	if delivered {
		t.Error("Deliver should return false when no clients connected")
	}
}
