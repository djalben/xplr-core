package logger

import (
	"log/slog"
	"strings"
)

const (
	Debug2            = -5              // Debug2 - самое подробное логирование, дампы пакетов, очень часто выполняемые операции, возможна повышенная нагрузка от логирования.
	Debug1 slog.Level = slog.LevelDebug // Debug1 - детальный лог. Например все выполняемые хттп-запросы с параметрами, но не настолько детальный, чтобы вызывать повышенную нагрузку от логов.
	Log    slog.Level = -1              // Log - обычный уровень логирования, рядовые операции дающие понимание того, чем занят сервис. По идее это нормальный уровень логирования для продакшена.
	Info              = slog.LevelInfo  // Info - важные вехи в работе сервиса, инициализация, завершение, рестарт отдельных субмодулей.
	Notice slog.Level = 2               // Notice - события требующие повышенного внимания.
	Warn              = slog.LevelWarn  // Warn - предупреждения, указывающие на некорректное поведение отдельных компонентов, но которое не является критичным для системы.
	Error             = slog.LevelError // Error - ошибка, сбой в работе компонента, который может повлиять на корректность работы сервиса.
	Panic  slog.Level = 10              // Panic - фатальная ошибка, аварийное завершение
)

// ToString - форматирование для админки.
func ToString(level slog.Level) string {
	switch level {
	case Debug2:
		return "D2"
	case Debug1:
		return "D1"
	case Log:
		return ""
	case Info:
		return "I"
	case Notice:
		return "N"
	case Warn:
		return "W"
	case Error:
		return "E"
	case Panic:
		return "!"
	default:
		return ""
	}
}

// GetLevel - получение нужного уровня из настроек.
func GetLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "panic":
		return Panic
	case "error":
		return Error
	case "warn":
		return Warn
	case "notice":
		return Notice
	case "info":
		return Info
	case "log":
		return Log
	case "debug1":
		return Debug1
	case "debug2":
		return Debug2

	default:
		return Log
	}
}
