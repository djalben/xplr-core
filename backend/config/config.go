package config

import (
	"os"
	"time"
)

// --- Константы CORE Logic ---

// MaxFailedAttempts - Количество отказов, после которого карта автоматически блокируется (Задача 2.4)
const MaxFailedAttempts = 3

// SuccessFeeRate - Комиссия за успешную транзакцию (например, 2%)
const SuccessFeeRate = 0.02

// CardIssuePrice - Стоимость выпуска одной карты (например, 5.00 USD)
const CardIssuePrice = 5.00

// --- Константы JWT ---

// TokenLifespan - Время жизни JWT-токена
const TokenLifespan = time.Hour * 24 * 7 // 7 дней

// --- Константы Базы Данных ---
// Рекомендуется читать из ENV, но для MVP можно оставить здесь
// const DatabaseURL = "postgresql://..."

// --- Telegram Configuration ---

// GetTelegramBotToken - Получить токен Telegram бота из ENV
func GetTelegramBotToken() string {
	return os.Getenv("TELEGRAM_BOT_TOKEN")
}

// GetTelegramAdminID - Получить ID администратора для уведомлений
func GetTelegramAdminID() string {
	adminID := os.Getenv("TELEGRAM_ADMIN_ID")
	if adminID == "" {
		// Fallback на TELEGRAM_CHAT_ID для обратной совместимости
		return os.Getenv("TELEGRAM_CHAT_ID")
	}
	return adminID
}

// GetTelegramChatID - Получить Chat ID для уведомлений (legacy)
func GetTelegramChatID() string {
	return os.Getenv("TELEGRAM_CHAT_ID")
}