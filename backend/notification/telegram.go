package notification

import (
	"fmt"
	"log"

	"github.com/aalabin/xplr/backend/telegram"
)

// SendTelegramMessage отправляет сообщение в Telegram.
// Если задан TELEGRAM_BOT_TOKEN (через telegram.SetBotToken в main), вызывается реальная отправка.
// Иначе — только лог (MVP-режим).
func SendTelegramMessage(chatID int64, message string) {
	if chatID == 0 {
		return
	}
	// Реальная отправка через пакет telegram (если токен задан в main)
	telegram.SendMessage(chatID, message)
	// Лог для отладки
	log.Printf("[TELEGRAM NOTIF] -> ChatID: %d | Message: %s", chatID, message)
}

// FormatAuthMessage форматирует сообщение для Telegram на основе результата авторизации.
// Это устраняет ошибки "undefined: notification.FormatAuthMessage"
func FormatAuthMessage(status, message string, cardID int, amount float64) string {
	emoji := "❌"
	if status == "APPROVED" {
		emoji = "✅"
	}
	
	// В MVP используем простой формат
	return fmt.Sprintf("%s Транзакция по карте %d:\nСтатус: %s\nСумма: %.2f\nСообщение: %s", 
		emoji, cardID, status, amount, message)
}