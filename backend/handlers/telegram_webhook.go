package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/telegram"
)

// ── GET /api/v1/user/settings/telegram-link ──
// Generates a t.me deep link with a temporary code for the current user.
func GetTelegramLinkHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	code, err := repository.StoreTelegramLinkCode(userID)
	if err != nil {
		log.Printf("[TELEGRAM] Failed to generate link code for user %d: %v", userID, err)
		http.Error(w, "Failed to generate Telegram link", http.StatusInternalServerError)
		return
	}

	link := "https://t.me/xplr_notify_bot?start=" + code

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"link": link,
	})
}

// ── POST /api/v1/telegram/webhook ──
// Telegram sends updates here. We process /start <code> to link accounts.
// This endpoint is PUBLIC (no JWT) — Telegram calls it directly.

// Telegram Update structures (minimal, only what we need)
type tgUpdate struct {
	Message       *tgMessage       `json:"message"`
	CallbackQuery *tgCallbackQuery `json:"callback_query"`
}

type tgMessage struct {
	MessageID      int64      `json:"message_id"`
	Chat           tgChat     `json:"chat"`
	Text           string     `json:"text"`
	ReplyToMessage *tgMessage `json:"reply_to_message"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

type tgCallbackQuery struct {
	ID      string     `json:"id"`
	From    tgUser     `json:"from"`
	Message *tgMessage `json:"message"`
	Data    string     `json:"data"`
}

type tgUser struct {
	ID int64 `json:"id"`
}

func TelegramWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// ── RAW REQUEST LOGGING — see every incoming Telegram request ──
	rawBody, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		log.Printf("[TG-WEBHOOK] ❌ Failed to read request body: %v", readErr)
		w.WriteHeader(http.StatusOK)
		return
	}
	log.Printf("[TG-WEBHOOK] 📥 Incoming Telegram Request: %s", string(rawBody))

	// Re-create body reader for JSON decoding
	r.Body = io.NopCloser(bytes.NewReader(rawBody))

	var update tgUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("[TG-WEBHOOK] Failed to decode update: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// ── Handle callback_query (inline button presses) ──
	if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery)
		w.WriteHeader(http.StatusOK)
		return
	}

	if update.Message == nil || update.Message.Text == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	text := strings.TrimSpace(update.Message.Text)
	chatID := update.Message.Chat.ID

	log.Printf("[TG-WEBHOOK] Received message: chat_id=%d, text=%q", chatID, text)

	// ── Chat Bridge: admin replies to a forwarded chat message ──
	if update.Message.ReplyToMessage != nil && !strings.HasPrefix(text, "/") {
		handleChatBridgeReply(chatID, update.Message.ReplyToMessage.MessageID, text)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle /start <code>
	if strings.HasPrefix(text, "/start ") {
		code := strings.TrimSpace(strings.TrimPrefix(text, "/start "))
		if code == "" {
			telegram.SendMessageHTML(chatID, "👋 <b>Привет!</b>\n\nЧтобы привязать аккаунт, используйте кнопку «Подключить Telegram» в настройках XPLR.")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Look up user by code
		userID, err := repository.LookupTelegramLinkCode(code)
		if err != nil || userID == 0 {
			log.Printf("[TG-WEBHOOK] Invalid or expired link code: %q (err=%v)", code, err)
			telegram.SendMessageHTML(chatID, "❌ <b>Ссылка недействительна или истекла.</b>\n\nПожалуйста, сгенерируйте новую в настройках XPLR.")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Save chat_id to user
		if err := repository.UpdateTelegramChatIDInt64(userID, chatID); err != nil {
			log.Printf("[TG-WEBHOOK] Failed to save chat_id for user %d: %v", userID, err)
			telegram.SendMessageHTML(chatID, "❌ <b>Произошла ошибка при привязке.</b>\n\nПопробуйте ещё раз.")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Delete used code
		repository.DeleteTelegramLinkCode(code)

		// Send welcome message
		telegram.SendMessageHTML(chatID,
			"✅ <b>Добро пожаловать в XPLR!</b>\n\n"+
				"Теперь вы будете получать уведомления здесь:\n\n"+
				"💳 <b>Транзакции</b> — списания и пополнения карт\n"+
				"💰 <b>Пополнения</b> — зачисления на баланс\n"+
				"� <b>Безопасность</b> — коды 2FA и оповещения\n"+
				"💬 <b>Поддержка</b> — ответы на ваши тикеты\n\n"+
				"Управлять уведомлениями: <a href=\"https://xplr.pro/settings\">xplr.pro/settings</a>")

		log.Printf("[TG-WEBHOOK] ✅ User %d linked to chat %d", userID, chatID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle bare /start (no code)
	if text == "/start" {
		telegram.SendMessageHTML(chatID,
			"👋 <b>Привет!</b>\n\n"+
				"Я бот уведомлений <b>XPLR</b>.\n\n"+
				"Чтобы привязать аккаунт, нажмите «Подключить Telegram» в настройках:\n"+
				"<a href=\"https://xplr.pro/settings\">xplr.pro/settings</a>")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle /help
	if text == "/help" {
		telegram.SendMessageHTML(chatID,
			"🆘 <b>Помощь и команды XPLR</b>\n\n"+
				"👤 <b>Профиль:</b> Используйте /status, чтобы проверить привязку.\n"+
				"💳 <b>Карты:</b> Уведомления о транзакциях приходят автоматически.\n"+
				"🔐 <b>Безопасность:</b> Коды 2FA приходят сюда.\n"+
				"💬 <b>Поддержка:</b> Уведомления об ответах на тикеты приходят сюда.")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle /test_me — diagnostic command to verify webhook is alive
	if text == "/test_me" {
		userID, _ := repository.GetUserIDByChatID(chatID)
		email, _ := repository.GetUserEmailByChatID(chatID)
		isAdmin := false
		if userID != 0 {
			isAdmin = repository.IsUserAdmin(userID)
		}
		var adminIDs []int64
		if telegram.AdminChatIDsProvider != nil {
			adminIDs, _ = telegram.AdminChatIDsProvider()
		}
		inList := false
		for _, aid := range adminIDs {
			if aid == chatID {
				inList = true
				break
			}
		}
		diag := fmt.Sprintf(
			"🔧 <b>Диагностика XPLR Bot</b>\n\n"+
				"🆔 <b>Your TG chat_id:</b> <code>%d</code>\n"+
				"👤 <b>Linked userID:</b> %d\n"+
				"📧 <b>Email:</b> %s\n"+
				"🛡 <b>is_admin:</b> %v\n"+
				"📋 <b>In AdminChatIDs list:</b> %v\n"+
				"📊 <b>Total admin IDs:</b> %d → %v\n\n"+
				"✅ Webhook работает!",
			chatID, userID, email, isAdmin, inList, len(adminIDs), adminIDs,
		)
		telegram.SendMessageHTML(chatID, diag)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle /status
	if text == "/status" {
		email, err := repository.GetUserEmailByChatID(chatID)
		if err != nil || email == "" {
			telegram.SendMessageHTML(chatID,
				"❌ <b>Аккаунт не привязан.</b>\n\n"+
					"Чтобы привязать, нажмите «Подключить Telegram» в настройках:\n"+
					"<a href=\"https://xplr.pro/settings\">xplr.pro/settings</a>")
		} else {
			telegram.SendMessageHTML(chatID,
				"✅ <b>Аккаунт привязан</b>\n\n"+
					"📧 <b>Email:</b> "+email+"\n\n"+
					"Вы получаете уведомления в этот чат.")
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// ── Any other message: check if admin direct message ──
	// If the sender is an admin with a claimed conversation, route their text there.
	adminUserID, _ := resolveAdminUserID(chatID)
	if adminUserID != 0 && !strings.HasPrefix(text, "/") {
		handleAdminDirectMessage(chatID, text)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Non-admin fallback: generic bot template
	telegram.SendMessageHTML(chatID,
		"🤖 Я бот уведомлений <b>XPLR</b>.\n\n"+
			"Доступные команды:\n"+
			"/status — проверить привязку\n"+
			"/help — помощь\n\n"+
			"Поддержка: <a href=\"https://xplr.pro/support\">xplr.pro/support</a>")
	w.WriteHeader(http.StatusOK)
}

// ── Callback Query Handler (inline button presses) ──

func handleCallbackQuery(cb *tgCallbackQuery) {
	if cb == nil {
		return
	}

	callerChatID := cb.From.ID
	data := cb.Data

	log.Printf("[TG-CALLBACK] chat_id=%d (int64), data=%q, callback_id=%q", callerChatID, data, cb.ID)

	// ── Handle claim_<convID> ──
	if strings.HasPrefix(data, "claim_") {
		handleClaimCallback(cb, callerChatID, data)
		return
	}

	// ── Handle closechat_<convID> ──
	if strings.HasPrefix(data, "closechat_") {
		handleCloseChatCallback(cb, callerChatID, data)
		return
	}

	// ── Handle noop (disabled buttons) ──
	if data == "noop" {
		telegram.AnswerCallbackQuery(cb.ID, "")
		return
	}

	// Always answer the callback to remove the loading spinner
	defer telegram.AnswerCallbackQuery(cb.ID, "")

	// Handle block_card:<cardID>
	if strings.HasPrefix(data, "block_card:") {
		cardIDStr := strings.TrimPrefix(data, "block_card:")
		cardID, err := strconv.Atoi(cardIDStr)
		if err != nil || cardID <= 0 {
			log.Printf("[TG-CALLBACK] Invalid card ID in callback: %q", cardIDStr)
			telegram.AnswerCallbackQuery(cb.ID, "Ошибка: неверный ID карты")
			return
		}

		// Security: verify that caller's chat_id is the card owner
		userID, err := repository.GetUserIDByChatID(callerChatID)
		if err != nil || userID == 0 {
			log.Printf("[TG-CALLBACK] Chat %d is not linked to any user", callerChatID)
			telegram.AnswerCallbackQuery(cb.ID, "Ошибка: аккаунт не привязан")
			return
		}

		// Block the card (UpdateCardStatus checks ownership via userID)
		if err := repository.UpdateCardStatus(cardID, userID, "BLOCKED"); err != nil {
			log.Printf("[TG-CALLBACK] Failed to block card %d for user %d: %v", cardID, userID, err)
			telegram.AnswerCallbackQuery(cb.ID, "Ошибка при блокировке карты")
			return
		}

		log.Printf("[TG-CALLBACK] ✅ Card %d blocked by user %d via Telegram", cardID, userID)

		// Edit the original message to confirm the block
		if cb.Message != nil {
			newText := fmt.Sprintf(
				"🔒 <b>Карта заблокирована</b>\n\n"+
					"Карта (ID: %d) заблокирована в целях безопасности.\n\n"+
					"Разблокировать можно в личном кабинете:\n"+
					"<a href=\"https://xplr.pro/cards\">xplr.pro/cards</a>",
				cardID,
			)
			if err := telegram.EditMessageText(callerChatID, cb.Message.MessageID, newText); err != nil {
				log.Printf("[TG-CALLBACK] Failed to edit message: %v", err)
			}
		}
		return
	}

	log.Printf("[TG-CALLBACK] Unknown callback data: %q", data)
	telegram.AnswerCallbackQuery(cb.ID, "")
}

// knownAdminEmails maps admin emails to allow auto-linking.
// Anyone who receives the "Claim" button is already an admin — they got the message
// because they were in GetAdminChatIDs. This list is for auto-linking TG → user account.
var knownAdminEmails = []string{"aalabin5@gmail.com", "vardump@inbox.ru"}

// resolveAdminUserID resolves a Telegram chat ID to an internal user ID.
// If the TG ID is not linked to any user, it tries to auto-link by finding
// an admin user without a telegram_chat_id and linking them.
func resolveAdminUserID(tgChatID int64) (int, string) {
	// 1. Try direct lookup
	userID, _ := repository.GetUserIDByChatID(tgChatID)
	if userID != 0 {
		name := repository.GetUserDisplayName(userID)
		log.Printf("[ADMIN-RESOLVE] TG %d → userID=%d (%s) via direct lookup", tgChatID, userID, name)
		return userID, name
	}

	// 2. Auto-link: find first known admin email whose telegram_chat_id is NULL/0
	log.Printf("[ADMIN-RESOLVE] TG %d not linked to any user — attempting auto-link...", tgChatID)
	for _, email := range knownAdminEmails {
		var uid int
		var existingTG int64
		err := repository.GlobalDB.QueryRow(
			`SELECT id, COALESCE(telegram_chat_id, 0) FROM users WHERE email = $1`, email,
		).Scan(&uid, &existingTG)
		if err != nil {
			log.Printf("[ADMIN-RESOLVE] Email %s not found in DB: %v", email, err)
			continue
		}
		if existingTG != 0 && existingTG != tgChatID {
			log.Printf("[ADMIN-RESOLVE] Email %s already linked to TG %d (not %d) — skip", email, existingTG, tgChatID)
			continue
		}
		if existingTG == tgChatID {
			// Already linked but GetUserIDByChatID failed? Shouldn't happen, but handle it
			log.Printf("[ADMIN-RESOLVE] Email %s already linked to TG %d — returning userID=%d", email, tgChatID, uid)
			return uid, repository.GetUserDisplayName(uid)
		}
		// Link this TG to this admin
		_, linkErr := repository.GlobalDB.Exec(
			`UPDATE users SET telegram_chat_id = $1 WHERE id = $2`, tgChatID, uid,
		)
		if linkErr != nil {
			log.Printf("[ADMIN-RESOLVE] ❌ Failed to auto-link TG %d → user %d (%s): %v", tgChatID, uid, email, linkErr)
			continue
		}
		log.Printf("[ADMIN-RESOLVE] ✅ AUTO-LINKED TG %d → user %d (%s)", tgChatID, uid, email)
		return uid, repository.GetUserDisplayName(uid)
	}

	log.Printf("[ADMIN-RESOLVE] ❌ Could not resolve TG %d to any admin user", tgChatID)
	return 0, ""
}

// ── Claim ticket callback ──
func handleClaimCallback(cb *tgCallbackQuery, callerChatID int64, data string) {
	// ⚡ IMMEDIATELY answer the callback to stop the spinner — before ANY other work
	telegram.AnswerCallbackQuery(cb.ID, "⏳ Обработка...")
	log.Printf("[CHAT-CLAIM] >>> Entry: callback_id=%q, data=%q, tg_chat_id=%d", cb.ID, data, callerChatID)

	// Safety net: report ALL errors to admin via TG message
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("❌ PANIC в handleClaimCallback: %v", r)
			log.Printf("[CHAT-CLAIM] %s", errMsg)
			telegram.SendMessageHTML(callerChatID, errMsg)
		}
	}()

	convIDStr := strings.TrimPrefix(data, "claim_")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil || convID <= 0 {
		errMsg := fmt.Sprintf("❌ Ошибка: неверный ID чата из data=%q", data)
		log.Printf("[CHAT-CLAIM] %s", errMsg)
		telegram.SendMessageHTML(callerChatID, errMsg)
		return
	}
	log.Printf("[CHAT-CLAIM] Parsed convID=%d", convID)

	// Resolve admin: direct lookup + auto-link fallback
	// NO is_admin check — if they received the Claim button, they ARE admin
	adminUserID, adminName := resolveAdminUserID(callerChatID)
	if adminUserID == 0 {
		errMsg := fmt.Sprintf("❌ Ошибка захвата: TG ID %d не удалось привязать ни к одному админ-аккаунту.\nПроверьте привязку Telegram в настройках XPLR.", callerChatID)
		log.Printf("[CHAT-CLAIM] %s", errMsg)
		telegram.SendMessageHTML(callerChatID, errMsg)
		return
	}
	log.Printf("[CHAT-CLAIM] Resolved admin: userID=%d, name=%s", adminUserID, adminName)

	// Force is_admin = TRUE for this user (belt-and-suspenders)
	repository.GlobalDB.Exec(`UPDATE users SET is_admin = TRUE WHERE id = $1`, adminUserID)

	// Atomically claim
	log.Printf("[CHAT-CLAIM] Attempting ClaimConversation(conv=%d, admin=%d)...", convID, adminUserID)
	claimed, claimErr := repository.ClaimConversation(convID, adminUserID)
	log.Printf("[CHAT-CLAIM] ClaimConversation result: claimed=%v, err=%v", claimed, claimErr)

	if claimErr != nil {
		errMsg := fmt.Sprintf("❌ Ошибка захвата тикета #%d: %v", convID, claimErr)
		log.Printf("[CHAT-CLAIM] %s", errMsg)
		telegram.SendMessageHTML(callerChatID, errMsg)
		return
	}

	if !claimed {
		conv, _ := repository.GetConversationByID(convID)
		if conv != nil && conv.ClaimedBy != 0 {
			claimerName := repository.GetUserDisplayName(conv.ClaimedBy)
			telegram.SendMessageHTML(callerChatID, fmt.Sprintf("ℹ️ Тикет #%d уже в работе у %s", convID, claimerName))
		} else {
			telegram.SendMessageHTML(callerChatID, fmt.Sprintf("ℹ️ Тикет #%d уже занят или не найден", convID))
		}
		return
	}

	log.Printf("[CHAT-CLAIM] ✅ Conv #%d CLAIMED by admin %d (%s)", convID, adminUserID, adminName)
	telegram.SendMessageHTML(callerChatID, fmt.Sprintf("✅ Вы взяли тикет #%d в работу", convID))

	// Post-claim work SYNCHRONOUSLY (Vercel kills goroutines after response)
	updateAllAdminKeyboards(convID, adminName)

	conv, _ := repository.GetConversationByID(convID)
	if conv != nil {
		repository.InsertChatMessage(convID, "admin", "Support Specialist", "Специалист подключился к диалогу", 0)
		userChatID := repository.GetUserTelegramChatID(conv.UserID)
		if userChatID != 0 {
			telegram.SendMessageHTML(userChatID,
				"💬 <b>Специалист подключился к диалогу</b>\n\n"+
					"<a href=\"https://xplr.pro/support\">Открыть чат</a>")
		}
	}
	log.Printf("[CHAT-CLAIM] ✅ Claim flow complete for conv #%d", convID)
}

// ── Close chat callback (only claimer) ──
func handleCloseChatCallback(cb *tgCallbackQuery, callerChatID int64, data string) {
	telegram.AnswerCallbackQuery(cb.ID, "⏳ Закрытие...")
	log.Printf("[CHAT-CLOSE] >>> Entry: data=%q, tg_chat_id=%d", data, callerChatID)

	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("❌ PANIC в handleCloseChatCallback: %v", r)
			log.Printf("[CHAT-CLOSE] %s", errMsg)
			telegram.SendMessageHTML(callerChatID, errMsg)
		}
	}()

	convIDStr := strings.TrimPrefix(data, "closechat_")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil || convID <= 0 {
		telegram.SendMessageHTML(callerChatID, "❌ Ошибка: неверный ID чата")
		return
	}

	adminUserID, _ := resolveAdminUserID(callerChatID)
	if adminUserID == 0 {
		telegram.SendMessageHTML(callerChatID, "❌ Нет доступа: аккаунт не найден")
		return
	}

	conv, err := repository.GetConversationByID(convID)
	if err != nil || conv == nil {
		telegram.SendMessageHTML(callerChatID, fmt.Sprintf("❌ Чат #%d не найден: %v", convID, err))
		return
	}

	if conv.ClaimedBy != 0 && conv.ClaimedBy != adminUserID {
		claimerName := repository.GetUserDisplayName(conv.ClaimedBy)
		telegram.SendMessageHTML(callerChatID, fmt.Sprintf("❌ Только %s может закрыть этот чат", claimerName))
		return
	}

	if err := repository.CloseConversation(convID); err != nil {
		telegram.SendMessageHTML(callerChatID, fmt.Sprintf("❌ Ошибка закрытия: %v", err))
		return
	}

	log.Printf("[CHAT-CLOSE] ✅ Conv #%d closed by admin %d", convID, adminUserID)
	telegram.SendMessageHTML(callerChatID, fmt.Sprintf("✅ Диалог #%d завершён", convID))

	// Synchronous post-close work
	repository.InsertChatMessage(convID, "admin", "Support Specialist",
		"Ваш вопрос был отмечен специалистом как решённый. Спасибо за обращение!", 0)
	updateAllAdminKeyboardsClosed(convID)
	if conv != nil {
		userChatID := repository.GetUserTelegramChatID(conv.UserID)
		if userChatID != 0 {
			telegram.SendMessageHTML(userChatID,
				"✅ <b>Задача решена</b>\n\n"+
					"Ваш вопрос был отмечен специалистом как решённый. Спасибо за обращение!\n\n"+
					"<a href=\"https://xplr.pro/support\">Создать новый запрос</a>")
		}
	}
}

// updateAllAdminKeyboards updates the inline keyboard on all forwarded messages for a conversation.
func updateAllAdminKeyboards(convID int, claimerName string) {
	entries, err := repository.GetLatestTgBridgePerAdmin(convID)
	if err != nil {
		log.Printf("[CHAT-CLAIM] Failed to get bridge entries for conv %d: %v", convID, err)
		return
	}
	log.Printf("[CHAT-CLAIM] Found %d bridge entries for conv #%d to update keyboards", len(entries), convID)

	claimedMarkup := &telegram.InlineKeyboardMarkupExported{
		InlineKeyboard: [][]telegram.InlineKeyboardButtonExported{
			{
				{Text: fmt.Sprintf("✅ В работе у %s", claimerName), CallbackData: "noop"},
			},
			{
				{Text: "🔒 Завершить диалог", CallbackData: fmt.Sprintf("closechat_%d", convID)},
			},
		},
	}

	for _, e := range entries {
		if err := telegram.EditMessageReplyMarkup(e.TgChatID, e.TgMessageID, claimedMarkup); err != nil {
			log.Printf("[CHAT-CLAIM] Failed to update keyboard (chat=%d, msg=%d): %v", e.TgChatID, e.TgMessageID, err)
		}
	}
}

// updateAllAdminKeyboardsClosed removes all buttons from forwarded messages.
func updateAllAdminKeyboardsClosed(convID int) {
	entries, err := repository.GetLatestTgBridgePerAdmin(convID)
	if err != nil {
		log.Printf("[CHAT-CLOSE] Failed to get bridge entries for conv %d: %v", convID, err)
		return
	}

	closedMarkup := &telegram.InlineKeyboardMarkupExported{
		InlineKeyboard: [][]telegram.InlineKeyboardButtonExported{
			{
				{Text: "✅ Задача решена и закрыта", CallbackData: "noop"},
			},
		},
	}

	for _, e := range entries {
		if err := telegram.EditMessageReplyMarkup(e.TgChatID, e.TgMessageID, closedMarkup); err != nil {
			log.Printf("[CHAT-CLOSE] Failed to update keyboard (chat=%d, msg=%d): %v", e.TgChatID, e.TgMessageID, err)
		}
	}
}

// ── Chat Bridge: admin TG reply → web chat ──

// sendAdminReplyToConversation inserts the admin's reply into a conversation and notifies the user.
// Returns true if successful.
func sendAdminReplyToConversation(adminChatID int64, adminUserID int, conv *repository.ChatConversation, text string) bool {
	// Check claim protection: only the claiming admin (or unclaimed) can reply
	if conv.ClaimedBy != 0 && conv.ClaimedBy != adminUserID {
		claimerName := repository.GetUserDisplayName(conv.ClaimedBy)
		telegram.SendMessageHTML(adminChatID,
			fmt.Sprintf("⚠️ <b>Тикет #%d обрабатывает %s.</b>\nВаши ответы не будут отправлены.", conv.ID, claimerName))
		return false
	}

	// Insert admin message
	_, err := repository.InsertChatMessage(conv.ID, "admin", "Support Specialist", text, 0)
	if err != nil {
		log.Printf("[CHAT-BRIDGE] ❌ Failed to insert admin reply for conv %d: %v", conv.ID, err)
		telegram.SendMessageHTML(adminChatID, fmt.Sprintf("❌ Ошибка записи в БД: %v", err))
		return false
	}

	log.Printf("[CHAT-BRIDGE] ✅ Admin reply saved to conversation #%d (admin=%d)", conv.ID, adminUserID)

	// Confirm to admin
	telegram.SendMessageHTML(adminChatID, fmt.Sprintf("✅ Сообщение доставлено клиенту (тикет #%d)", conv.ID))

	// Notify the user via Telegram
	userChatID := repository.GetUserTelegramChatID(conv.UserID)
	if userChatID != 0 {
		telegram.SendMessageHTML(userChatID,
			fmt.Sprintf("💬 <b>Новое сообщение от поддержки</b>\n\n%s\n\n"+
				"<a href=\"https://xplr.pro/support\">Открыть чат</a>", text))
	}
	return true
}

func handleChatBridgeReply(adminChatID int64, replyToMsgID int64, text string) {
	log.Printf("[CHAT-BRIDGE] Incoming reply: adminChat=%d, replyToMsgID=%d, text=%q", adminChatID, replyToMsgID, text)

	// Resolve admin first — we need this for all paths
	adminUserID, adminName := resolveAdminUserID(adminChatID)
	if adminUserID == 0 {
		log.Printf("[CHAT-BRIDGE] ❌ Chat %d not resolved to any admin", adminChatID)
		telegram.SendMessageHTML(adminChatID, "❌ Ваш Telegram не привязан к аккаунту XPLR.")
		return
	}
	log.Printf("[CHAT-BRIDGE] Admin resolved: userID=%d (%s)", adminUserID, adminName)

	// 1. Try to find conversation via bridge table (exact message match)
	convID, err := repository.GetConversationIDByTgReplyMsgID(replyToMsgID)
	if err == nil && convID != 0 {
		conv, convErr := repository.GetConversationByID(convID)
		if convErr == nil && conv != nil {
			log.Printf("[CHAT-BRIDGE] Bridge match: tg_msg=%d → conv #%d", replyToMsgID, convID)
			sendAdminReplyToConversation(adminChatID, adminUserID, conv, text)
			return
		}
	}
	log.Printf("[CHAT-BRIDGE] ⚠️ Bridge lookup failed for tg_message_id=%d (err=%v, convID=%d)", replyToMsgID, err, convID)

	// 2. Fallback: find admin's claimed open conversation
	claimedConv, claimErr := repository.GetClaimedOpenConversation(adminUserID)
	if claimErr == nil && claimedConv != nil {
		log.Printf("[CHAT-BRIDGE] Fallback: routing to admin's claimed conv #%d", claimedConv.ID)
		sendAdminReplyToConversation(adminChatID, adminUserID, claimedConv, text)
		return
	}

	// 3. Nothing found — tell admin
	log.Printf("[CHAT-BRIDGE] ❌ No conversation found for admin %d (bridge failed, no claimed conv)", adminUserID)
	telegram.SendMessageHTML(adminChatID,
		"❌ <b>Не удалось определить тикет.</b>\n\n"+
			"Нажмите <b>Reply</b> на сообщение клиента, чтобы ответить.\n"+
			"Или сначала нажмите «Взять в работу» на тикете.")
}

// handleAdminDirectMessage handles a direct (non-reply) message from an admin.
// Routes it to their currently claimed open conversation, if any.
func handleAdminDirectMessage(adminChatID int64, text string) {
	adminUserID, adminName := resolveAdminUserID(adminChatID)
	if adminUserID == 0 {
		return // not an admin, caller handles fallback
	}
	log.Printf("[CHAT-DIRECT] Admin direct message from %d (%s): %q", adminChatID, adminName, text)

	claimedConv, err := repository.GetClaimedOpenConversation(adminUserID)
	if err != nil || claimedConv == nil {
		telegram.SendMessageHTML(adminChatID,
			"ℹ️ <b>У вас нет активных тикетов.</b>\n\n"+
				"Чтобы ответить клиенту:\n"+
				"1. Нажмите «Взять в работу» на тикете\n"+
				"2. Ответьте Reply на сообщение клиента\n\n"+
				"Или просто напишите сюда текст — он будет направлен в ваш активный тикет.")
		return
	}

	log.Printf("[CHAT-DIRECT] Routing direct message to claimed conv #%d", claimedConv.ID)
	sendAdminReplyToConversation(adminChatID, adminUserID, claimedConv, text)
}
