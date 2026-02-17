package telegram

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var botToken string
// var chatID string // УДАЛЕНО: Глобальный ChatID больше не нужен

// SetBotToken устанавливает токен бота.
func SetBotToken(token string) {
	botToken = token
}

// SetChatID (заглушка) - Теперь не используется, ChatID берется из БД.
func SetChatID(id string) {
	log.Printf("DEPRECATED: Global SetChatID called with ID: %s. This function should be removed from main.go.", id)
}


// NotifyDepositToChat отправляет уведомление о пополнении через API Telegram на указанный ChatID.
// Принимает ChatID как int64, так как он приходит из БД.
func NotifyDepositToChat(chatID int64, userID int, amount float64, newBalance float64) {
	if botToken == "" {
		log.Println("Telegram notification skipped: Bot token is not set.")
		return
	}
    
    // Преобразование ChatID (int64) в строку для URL
    chatIDStr := fmt.Sprintf("%d", chatID)

	// 1. Формирование сообщения
	message := fmt.Sprintf(
		"💸 НОВЫЙ ДЕПОЗИТ (User %d)\n\nСумма: %.2f EUR\nНовый баланс: %.2f EUR",
		userID, 
		amount, 
		newBalance,
	)

	// 2. Кодирование сообщения для URL
	encodedMessage := url.QueryEscape(message)

	// 3. Формирование URL для API Telegram
	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
		botToken, 
		chatIDStr, // ИСПОЛЬЗУЕМ ЛОКАЛЬНЫЙ chatIDStr
		encodedMessage,
	)
	
	// 4. Отправка HTTP-запроса
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Telegram notify failed (User %d, Chat %d): HTTP request error: %v", userID, chatID, err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram notify failed (User %d, Chat %d): API returned status %d", userID, chatID, resp.StatusCode)
		return
	}
	
	log.Printf("Telegram deposit notification sent successfully to chat %d for user %d.", chatID, userID)
}


// NotifyDeposit - Старая функция, которую нужно удалить.
func NotifyDeposit(userID int, amount float64, newBalance float64) {
	log.Println("DEPRECATED: Old NotifyDeposit called. Action skipped.")
}

// SendMessage отправляет произвольное текстовое сообщение в Telegram (если задан bot token).
// Используется из notification пакета для депозитов, блокировки карты и т.д.
func SendMessage(chatID int64, message string) {
	if botToken == "" {
		return
	}
	if chatID == 0 {
		return
	}
	chatIDStr := fmt.Sprintf("%d", chatID)
	encodedMessage := url.QueryEscape(message)
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
		botToken, chatIDStr, encodedMessage)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Telegram SendMessage failed (Chat %d): %v", chatID, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram SendMessage failed (Chat %d): API returned %d", chatID, resp.StatusCode)
	}
}