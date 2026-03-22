package config

import (
	"gitlab.com/libs-artifex/envparse"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// ENV — основная конфигурация проекта (берётся из .env + окружения).
type ENV struct {
	// База данных
	PostgresDSN string `env:"POSTGRES_DSN" required:"true"`

	// HTTP сервер
	ServerHost string `env:"SERVER_HOST" default:"0.0.0.0"`
	ServerPort int    `env:"SERVER_PORT" default:"8080"`

	// JWT
	JWTSecret string `env:"JWT_SECRET" required:"true"`

	// Логи
	LogLevel string `env:"LOG_LEVEL" default:"info"`
	LogPlain bool   `env:"LOG_PLAIN" default:"false"`

	// Telegram (для тикетов и уведомлений)
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN"`
	TelegramChatID   int64  `env:"TELEGRAM_CHAT_ID"` // основной чат поддержки

	// Эмитент карт (потом добавим)
	CardEmitterAPIKey string `env:"CARD_EMITTER_API_KEY"`
	CardEmitterURL    string `env:"CARD_EMITTER_URL"`

	// Дополнительно
	Debug bool `env:"DEBUG" default:"false"`
}

// Parse — загрузка конфига из окружения + .env файла.
func Parse() (ENV, error) {
	cfg := ENV{}

	err := envparse.Process("", &cfg)
	if err != nil {
		return ENV{}, wrapper.Wrap(err)
	}

	return cfg, nil
}
