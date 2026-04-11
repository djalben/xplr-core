package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// SendMessageHTML отправляет сообщение с HTML-разметкой (parse_mode=HTML).
func SendMessageHTML(chatID int64, message string) {
	if botToken == "" {
		log.Printf("[TELEGRAM] ⚠️ SendMessageHTML skipped: TELEGRAM_BOT_TOKEN not set (chat_id=%d)", chatID)
		return
	}
	if chatID == 0 {
		log.Printf("[TELEGRAM] ⚠️ SendMessageHTML skipped: chatID is 0")
		return
	}
	chatIDStr := fmt.Sprintf("%d", chatID)
	encodedMessage := url.QueryEscape(message)
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&parse_mode=HTML&text=%s",
		botToken, chatIDStr, encodedMessage)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("[TELEGRAM] ❌ SendMessageHTML failed (Chat %d): %v", chatID, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[TELEGRAM] ❌ SendMessageHTML failed (Chat %d): API status %d, body: %s", chatID, resp.StatusCode, string(body))
	}
}

// SendMessageHTMLSafe — error-returning version of SendMessageHTML for use in NotifyUser.
// Uses POST with JSON body for reliability (no URL length limits, proper encoding).
func SendMessageHTMLSafe(chatID int64, message string) error {
	if botToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN not set — Telegram notifications disabled")
	}
	if chatID == 0 {
		return fmt.Errorf("chatID is 0 — cannot send Telegram message")
	}

	type sendMsgPayload struct {
		ChatID    int64  `json:"chat_id"`
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}

	// Telegram sendMessage text limit is 4096 chars
	text := message
	if len(text) > 4000 {
		text = text[:4000] + "..."
	}

	if err := postJSON("sendMessage", sendMsgPayload{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}); err != nil {
		log.Printf("[TELEGRAM] ❌ SendMessageHTMLSafe POST failed for chat %d: %v", chatID, err)
		return err
	}
	return nil
}

// NotifyAdmins отправляет HTML-сообщение всем админам с привязанным Telegram.
// Если buttonText и buttonURL не пустые — добавляет inline URL-кнопку.
// adminChatIDs — функция-источник chat_id (инъекция зависимости, чтобы не импортировать repository).
var AdminChatIDsProvider func() ([]int64, error)

func NotifyAdmins(message string, buttonText string, buttonURL string) {
	if botToken == "" {
		log.Println("[TELEGRAM] ⚠️ NotifyAdmins skipped: TELEGRAM_BOT_TOKEN not set")
		return
	}
	if AdminChatIDsProvider == nil {
		log.Println("[TELEGRAM] ⚠️ NotifyAdmins skipped: AdminChatIDsProvider not set (check main.go init)")
		return
	}
	ids, err := AdminChatIDsProvider()
	if err != nil {
		log.Printf("[TELEGRAM] ❌ NotifyAdmins: failed to get admin chat IDs: %v", err)
		return
	}
	if len(ids) == 0 {
		log.Println("[TELEGRAM] ⚠️ NotifyAdmins: no admins with linked Telegram found (telegram_chat_id is NULL for all admins)")
		return
	}

	log.Printf("[TELEGRAM] 📤 Sending notification to %d admin(s)...", len(ids))
	for _, chatID := range ids {
		if buttonText != "" && buttonURL != "" {
			payload := sendMessageWithURLPayload{
				ChatID:    chatID,
				Text:      message,
				ParseMode: "HTML",
				ReplyMarkup: &inlineKeyboardURLMarkup{
					InlineKeyboard: [][]inlineKeyboardURLButton{
						{
							{Text: buttonText, URL: buttonURL},
						},
					},
				},
			}
			if err := postJSON("sendMessage", payload); err != nil {
				log.Printf("[TELEGRAM] NotifyAdmins failed (Chat %d): %v", chatID, err)
			}
		} else {
			SendMessageHTML(chatID, message)
		}
	}
}

// ── Inline Keyboard helpers ──

type inlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type sendMessagePayload struct {
	ChatID      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ParseMode   string                `json:"parse_mode,omitempty"`
	ReplyMarkup *inlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type inlineKeyboardMarkup struct {
	InlineKeyboard [][]inlineKeyboardButton `json:"inline_keyboard"`
}

// URL button types (for NotifyAdmins — opens a link, no callback)
type inlineKeyboardURLButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type inlineKeyboardURLMarkup struct {
	InlineKeyboard [][]inlineKeyboardURLButton `json:"inline_keyboard"`
}

type sendMessageWithURLPayload struct {
	ChatID      int64                    `json:"chat_id"`
	Text        string                   `json:"text"`
	ParseMode   string                   `json:"parse_mode,omitempty"`
	ReplyMarkup *inlineKeyboardURLMarkup `json:"reply_markup,omitempty"`
}

type editMessagePayload struct {
	ChatID      int64                 `json:"chat_id"`
	MessageID   int64                 `json:"message_id"`
	Text        string                `json:"text"`
	ParseMode   string                `json:"parse_mode,omitempty"`
	ReplyMarkup *inlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type answerCallbackPayload struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text,omitempty"`
	ShowAlert       bool   `json:"show_alert,omitempty"`
}

// postJSON sends a JSON POST to the Telegram Bot API.
func postJSON(method string, payload interface{}) error {
	if botToken == "" {
		return fmt.Errorf("bot token not set")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API %s returned %d: %s", method, resp.StatusCode, string(respBody))
	}
	return nil
}

// SendTransactionAlert отправляет уведомление о транзакции с кнопкой блокировки карты.
// cardID используется в callback_data для идентификации карты.
func SendTransactionAlert(chatID int64, amount string, merchant string, last4 string, balance string, cardID int) {
	if botToken == "" || chatID == 0 {
		return
	}

	text := fmt.Sprintf(
		"🔴 <b>Попытка списания</b>\n\n"+
			"💰 <b>Сумма:</b> %s\n"+
			"🏪 <b>Мерчант:</b> %s\n"+
			"💳 <b>Карта:</b> •••• %s\n"+
			"💵 <b>Баланс:</b> %s",
		amount, merchant, last4, balance,
	)

	payload := sendMessagePayload{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
		ReplyMarkup: &inlineKeyboardMarkup{
			InlineKeyboard: [][]inlineKeyboardButton{
				{
					{
						Text:         "❌ ЗАБЛОКИРОВАТЬ КАРТУ",
						CallbackData: fmt.Sprintf("block_card:%d", cardID),
					},
				},
			},
		},
	}

	if err := postJSON("sendMessage", payload); err != nil {
		log.Printf("[TELEGRAM] SendTransactionAlert failed (Chat %d, Card %d): %v", chatID, cardID, err)
	} else {
		log.Printf("[TELEGRAM] Transaction alert sent to chat %d for card %d", chatID, cardID)
	}
}

// SendMessageHTMLReturnID sends an HTML message and returns the Telegram message_id.
// Used by the chat bridge to save the TG message ID for reply routing.
func SendMessageHTMLReturnID(chatID int64, message string) int64 {
	if botToken == "" || chatID == 0 {
		return 0
	}
	payload := sendMessagePayload{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "HTML",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[TELEGRAM] SendMessageHTMLReturnID marshal error: %v", err)
		return 0
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[TELEGRAM] SendMessageHTMLReturnID HTTP error (Chat %d): %v", chatID, err)
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[TELEGRAM] SendMessageHTMLReturnID failed (Chat %d): %d %s", chatID, resp.StatusCode, string(respBody))
		return 0
	}
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Try reading the body we already consumed — fallback
		return 0
	}
	return result.Result.MessageID
}

// EditMessageText изменяет текст существующего сообщения (убирает inline-кнопки).
func EditMessageText(chatID int64, messageID int64, newText string) error {
	payload := editMessagePayload{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      newText,
		ParseMode: "HTML",
	}
	return postJSON("editMessageText", payload)
}

// editReplyMarkupPayload is used by editMessageReplyMarkup to update only the keyboard.
type editReplyMarkupPayload struct {
	ChatID      int64       `json:"chat_id"`
	MessageID   int64       `json:"message_id"`
	ReplyMarkup interface{} `json:"reply_markup"`
}

// EditMessageReplyMarkup updates only the inline keyboard of an existing message.
func EditMessageReplyMarkup(chatID int64, messageID int64, markup interface{}) error {
	payload := editReplyMarkupPayload{
		ChatID:      chatID,
		MessageID:   messageID,
		ReplyMarkup: markup,
	}
	return postJSON("editMessageReplyMarkup", payload)
}

// SendMessageHTMLWithKeyboardReturnID sends an HTML message with an inline keyboard and returns the message_id.
func SendMessageHTMLWithKeyboardReturnID(chatID int64, message string, keyboard *inlineKeyboardMarkup) int64 {
	if botToken == "" || chatID == 0 {
		return 0
	}
	payload := sendMessagePayload{
		ChatID:      chatID,
		Text:        message,
		ParseMode:   "HTML",
		ReplyMarkup: keyboard,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[TELEGRAM] SendMessageHTMLWithKeyboardReturnID marshal error: %v", err)
		return 0
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[TELEGRAM] SendMessageHTMLWithKeyboardReturnID HTTP error (Chat %d): %v", chatID, err)
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[TELEGRAM] SendMessageHTMLWithKeyboardReturnID failed (Chat %d): %d %s", chatID, resp.StatusCode, string(respBody))
		return 0
	}
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0
	}
	return result.Result.MessageID
}

// BuildInlineKeyboard is a helper to create an inline keyboard markup from rows of buttons.
func BuildInlineKeyboard(rows [][]inlineKeyboardButton) *inlineKeyboardMarkup {
	return &inlineKeyboardMarkup{InlineKeyboard: rows}
}

// NewCallbackButton creates a single inline keyboard button with callback_data.
func NewCallbackButton(text string, callbackData string) inlineKeyboardButton {
	return inlineKeyboardButton{Text: text, CallbackData: callbackData}
}

// ── Exported types for cross-package use (handlers → telegram) ──

// InlineKeyboardButtonExported is an exported inline keyboard button.
type InlineKeyboardButtonExported struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

// InlineKeyboardMarkupExported is an exported inline keyboard markup.
type InlineKeyboardMarkupExported struct {
	InlineKeyboard [][]InlineKeyboardButtonExported `json:"inline_keyboard"`
}

// SendMessageHTMLWithInlineReturnID sends an HTML message with an inline keyboard and returns the message_id.
func SendMessageHTMLWithInlineReturnID(chatID int64, message string, keyboard *InlineKeyboardMarkupExported) int64 {
	if botToken == "" || chatID == 0 {
		return 0
	}
	type payload struct {
		ChatID      int64                         `json:"chat_id"`
		Text        string                        `json:"text"`
		ParseMode   string                        `json:"parse_mode,omitempty"`
		ReplyMarkup *InlineKeyboardMarkupExported `json:"reply_markup,omitempty"`
	}
	p := payload{ChatID: chatID, Text: message, ParseMode: "HTML", ReplyMarkup: keyboard}
	body, err := json.Marshal(p)
	if err != nil {
		log.Printf("[TELEGRAM] SendMessageHTMLWithInlineReturnID marshal error: %v", err)
		return 0
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[TELEGRAM] SendMessageHTMLWithInlineReturnID HTTP error (Chat %d): %v", chatID, err)
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[TELEGRAM] SendMessageHTMLWithInlineReturnID failed (Chat %d): %d %s", chatID, resp.StatusCode, string(respBody))
		return 0
	}
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0
	}
	return result.Result.MessageID
}

// SendPhotoWithCaption sends a photo by URL with an HTML caption via Telegram sendPhoto API.
// If photoURL is empty, falls back to SendMessageHTMLSafe (text-only).
// Caption is truncated to 1024 chars (Telegram sendPhoto limit).
func SendPhotoWithCaption(chatID int64, photoURL string, caption string) error {
	if botToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN not set")
	}
	if chatID == 0 {
		return fmt.Errorf("chatID is 0")
	}
	if photoURL == "" {
		log.Printf("[TELEGRAM] SendPhotoWithCaption: no photo URL, falling back to text for chat %d", chatID)
		return SendMessageHTMLSafe(chatID, caption)
	}

	// Telegram sendPhoto caption limit is 1024 chars
	safeCaption := caption
	if len(safeCaption) > 1000 {
		safeCaption = safeCaption[:1000] + "..."
	}

	type sendPhotoPayload struct {
		ChatID    int64  `json:"chat_id"`
		Photo     string `json:"photo"`
		Caption   string `json:"caption"`
		ParseMode string `json:"parse_mode"`
	}

	payload := sendPhotoPayload{
		ChatID:    chatID,
		Photo:     photoURL,
		Caption:   safeCaption,
		ParseMode: "HTML",
	}

	if err := postJSON("sendPhoto", payload); err != nil {
		log.Printf("[TELEGRAM] ❌ SendPhotoWithCaption failed (Chat %d, photo=%s): %v — falling back to text", chatID, photoURL[:min(len(photoURL), 60)], err)
		return SendMessageHTMLSafe(chatID, caption)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AnswerCallbackQuery отвечает на callback_query (убирает «часики» в Telegram).
func AnswerCallbackQuery(callbackQueryID string, text string) {
	payload := answerCallbackPayload{
		CallbackQueryID: callbackQueryID,
		Text:            text,
		ShowAlert:       false,
	}
	if err := postJSON("answerCallbackQuery", payload); err != nil {
		log.Printf("[TELEGRAM] AnswerCallbackQuery failed: %v", err)
	}
}
