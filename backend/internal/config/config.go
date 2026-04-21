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
	// CORS: пусто = dev (AllowedOrigins * без credentials). В проде — список через запятую, например https://app.example.com
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS"`

	// JWT
	JWTSecret string `env:"JWT_SECRET_TOKEN" required:"true"`

	// Логи
	LogLevel string `env:"LOG_LEVEL" default:"info"`
	LogPlain bool   `env:"LOG_PLAIN" default:"false"`

	// Telegram (для тикетов и уведомлений)
	TelegramBotToken    string `env:"TELEGRAM_BOT_TOKEN"`
	TelegramChatID      int64  `env:"TELEGRAM_CHAT_ID"`                                // основной чат поддержки
	TelegramBotUsername string `env:"TELEGRAM_BOT_USERNAME" default:"xplr_notify_bot"` // для deep-link привязки в настройках

	// Эмитент карт (потом добавим)
	CardEmitterAPIKey string `env:"CARD_EMITTER_API_KEY"`
	CardEmitterURL    string `env:"CARD_EMITTER_URL"`

	// Дополнительно
	Debug bool `env:"DEBUG" default:"false"`

	// Ссылки в письмах (верификация, сброс пароля)
	AppPublicURL string `env:"APP_PUBLIC_URL" default:"http://localhost:8080"`

	// SMTP (пустой SMTP_HOST — письма не отправляются, только Noop mailer)
	SMTPHost     string `env:"SMTP_HOST"`
	SMTPPort     int    `env:"SMTP_PORT" default:"465"`
	SMTPUser     string `env:"SMTP_USER"`
	SMTPPassword string `env:"SMTP_PASS"`
	SMTPFrom     string `env:"SMTP_FROM"`

	// Store providers
	MobiMatterAPIKey string `env:"MOBIMATTER_API_KEY"`
	MobiMatterAPIURL string `env:"MOBIMATTER_API_URL" default:"https://api.mobimatter.com/mobimatter/api/v2"`

	// 3X-UI / XPanel (VLESS+Reality)
	XPanelURL              string `env:"XPANEL_URL"`
	XPanelBasePath         string `env:"XPANEL_BASE_PATH" default:"/panel"`
	XPanelUsername         string `env:"XPANEL_USERNAME"`
	XPanelPassword         string `env:"XPANEL_PASSWORD"`
	XPanelInboundID        int    `env:"XPANEL_INBOUND_ID" default:"1"`
	XPanelServerIP         string `env:"XPANEL_SERVER_IP"`
	XPanelServerPort       string `env:"XPANEL_SERVER_PORT" default:"443"`
	XPanelSNI              string `env:"XPANEL_SNI" default:"www.microsoft.com"`
	XPanelRealityPublicKey string `env:"XPANEL_REALITY_PUBLIC_KEY"`
	XPanelRealityShortID   string `env:"XPANEL_REALITY_SHORT_ID"`
	XPanelFlow             string `env:"XPANEL_FLOW" default:"xtls-rprx-vision"`
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
