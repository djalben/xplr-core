package handlers

import (
	"encoding/json"
	"fmt"
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
	var update tgUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("[TG-WEBHOOK] Failed to decode update: %v", err)
		// Always return 200 to Telegram so it doesn't retry
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

	// Any other message
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

	log.Printf("[TG-CALLBACK] chat_id=%d, data=%q", callerChatID, data)

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

// ── Claim ticket callback ──
func handleClaimCallback(cb *tgCallbackQuery, callerChatID int64, data string) {
	// Safety net: always answer callback to stop the spinner, even on panic
	answered := false
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CHAT-CLAIM] ❌ PANIC in handleClaimCallback: %v", r)
		}
		if !answered {
			telegram.AnswerCallbackQuery(cb.ID, "Ошибка обработки")
		}
	}()

	log.Printf("[CHAT-CLAIM] >>> Entry: callback_id=%q, data=%q, tg_user_id=%d", cb.ID, data, callerChatID)

	convIDStr := strings.TrimPrefix(data, "claim_")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil || convID <= 0 {
		log.Printf("[CHAT-CLAIM] ❌ Invalid convID from data=%q: parsed=%q err=%v", data, convIDStr, err)
		telegram.AnswerCallbackQuery(cb.ID, "Ошибка: неверный ID чата")
		answered = true
		return
	}
	log.Printf("[CHAT-CLAIM] Parsed convID=%d", convID)

	// Verify caller is an admin
	adminUserID, dbErr := repository.GetUserIDByChatID(callerChatID)
	log.Printf("[CHAT-CLAIM] GetUserIDByChatID(tg=%d) → userID=%d, err=%v", callerChatID, adminUserID, dbErr)
	if adminUserID == 0 {
		log.Printf("[CHAT-CLAIM] ❌ TG chat %d is not linked to any user", callerChatID)
		telegram.AnswerCallbackQuery(cb.ID, "Ошибка: аккаунт не привязан к XPLR")
		answered = true
		return
	}

	isAdmin := repository.IsUserAdmin(adminUserID)
	log.Printf("[CHAT-CLAIM] IsUserAdmin(userID=%d) → %v", adminUserID, isAdmin)
	if !isAdmin {
		log.Printf("[CHAT-CLAIM] ❌ User %d (tg=%d) is NOT admin", adminUserID, callerChatID)
		telegram.AnswerCallbackQuery(cb.ID, "Нет доступа: вы не администратор")
		answered = true
		return
	}

	// Atomically claim
	log.Printf("[CHAT-CLAIM] Attempting ClaimConversation(conv=%d, admin=%d)...", convID, adminUserID)
	claimed, claimErr := repository.ClaimConversation(convID, adminUserID)
	log.Printf("[CHAT-CLAIM] ClaimConversation result: claimed=%v, err=%v", claimed, claimErr)

	if claimErr != nil {
		log.Printf("[CHAT-CLAIM] ❌ DB error claiming conv %d: %v", convID, claimErr)
		telegram.AnswerCallbackQuery(cb.ID, fmt.Sprintf("Ошибка БД: %v", claimErr))
		answered = true
		return
	}

	if !claimed {
		conv, _ := repository.GetConversationByID(convID)
		if conv != nil && conv.ClaimedBy != 0 {
			claimerName := repository.GetUserDisplayName(conv.ClaimedBy)
			log.Printf("[CHAT-CLAIM] Conv #%d already claimed by user %d (%s)", convID, conv.ClaimedBy, claimerName)
			telegram.AnswerCallbackQuery(cb.ID, fmt.Sprintf("Уже в работе у %s", claimerName))
		} else {
			log.Printf("[CHAT-CLAIM] Conv #%d claim returned 0 rows (conv=%+v)", convID, conv)
			telegram.AnswerCallbackQuery(cb.ID, "Тикет уже занят или не найден")
		}
		answered = true
		return
	}

	adminName := repository.GetUserDisplayName(adminUserID)
	log.Printf("[CHAT-CLAIM] ✅ Conv #%d CLAIMED by admin %d (%s)", convID, adminUserID, adminName)

	// Answer callback FIRST to stop the spinner immediately
	telegram.AnswerCallbackQuery(cb.ID, "✅ Вы взяли тикет в работу")
	answered = true

	// Everything below is async / best-effort — spinner is already gone
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[CHAT-CLAIM] ❌ PANIC in async claim work: %v", r)
			}
		}()

		// Update inline keyboard on ALL admin messages for this conversation
		updateAllAdminKeyboards(convID, adminName)

		// Send system message to web chat
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
		log.Printf("[CHAT-CLAIM] Async claim work complete for conv #%d", convID)
	}()
}

// ── Close chat callback (only claimer) ──
func handleCloseChatCallback(cb *tgCallbackQuery, callerChatID int64, data string) {
	answered := false
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CHAT-CLOSE] ❌ PANIC in handleCloseChatCallback: %v", r)
		}
		if !answered {
			telegram.AnswerCallbackQuery(cb.ID, "Ошибка обработки")
		}
	}()

	log.Printf("[CHAT-CLOSE] >>> Entry: data=%q, tg_user_id=%d", data, callerChatID)

	convIDStr := strings.TrimPrefix(data, "closechat_")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil || convID <= 0 {
		telegram.AnswerCallbackQuery(cb.ID, "Ошибка: неверный ID чата")
		answered = true
		return
	}

	adminUserID, _ := repository.GetUserIDByChatID(callerChatID)
	log.Printf("[CHAT-CLOSE] GetUserIDByChatID(tg=%d) → userID=%d", callerChatID, adminUserID)
	if adminUserID == 0 || !repository.IsUserAdmin(adminUserID) {
		telegram.AnswerCallbackQuery(cb.ID, "Нет доступа")
		answered = true
		return
	}

	conv, err := repository.GetConversationByID(convID)
	if err != nil || conv == nil {
		log.Printf("[CHAT-CLOSE] Conv %d not found: err=%v", convID, err)
		telegram.AnswerCallbackQuery(cb.ID, "Чат не найден")
		answered = true
		return
	}

	// Only the claimer can close
	if conv.ClaimedBy != adminUserID {
		log.Printf("[CHAT-CLOSE] Admin %d tried to close conv #%d owned by %d", adminUserID, convID, conv.ClaimedBy)
		telegram.AnswerCallbackQuery(cb.ID, "Только назначенный специалист может закрыть чат")
		answered = true
		return
	}

	if err := repository.CloseConversation(convID); err != nil {
		log.Printf("[CHAT-CLOSE] ❌ Failed to close conv %d: %v", convID, err)
		telegram.AnswerCallbackQuery(cb.ID, "Ошибка при закрытии")
		answered = true
		return
	}

	log.Printf("[CHAT-CLOSE] ✅ Conv #%d closed by admin %d", convID, adminUserID)
	telegram.AnswerCallbackQuery(cb.ID, "✅ Диалог завершён")
	answered = true

	// Async: system message + keyboard updates + user notification
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[CHAT-CLOSE] ❌ PANIC in async close work: %v", r)
			}
		}()
		repository.InsertChatMessage(convID, "admin", "Support Specialist", "Диалог завершён специалистом", 0)
		updateAllAdminKeyboardsClosed(convID)
		userChatID := repository.GetUserTelegramChatID(conv.UserID)
		if userChatID != 0 {
			telegram.SendMessageHTML(userChatID,
				"🔒 <b>Диалог завершён</b>\n\nСпасибо за обращение!\n"+
					"<a href=\"https://xplr.pro/support\">Открыть новый чат</a>")
		}
	}()
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
				{Text: "🔒 Диалог завершён", CallbackData: "noop"},
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

func handleChatBridgeReply(adminChatID int64, replyToMsgID int64, text string) {
	log.Printf("[CHAT-BRIDGE] Incoming reply: adminChat=%d, replyToMsgID=%d, text=%q", adminChatID, replyToMsgID, text)

	// 1. Find the conversation by the TG message the admin replied to
	convID, err := repository.GetConversationIDByTgReplyMsgID(replyToMsgID)
	if err != nil || convID == 0 {
		log.Printf("[CHAT-BRIDGE] ❌ No conversation found for tg_message_id=%d (err=%v)", replyToMsgID, err)
		return
	}

	conv, err := repository.GetConversationByID(convID)
	if err != nil || conv == nil {
		log.Printf("[CHAT-BRIDGE] ❌ Conversation %d not found in DB (err=%v)", convID, err)
		return
	}
	log.Printf("[CHAT-BRIDGE] Found conversation #%d (user=%d, topic=%q, status=%q)", conv.ID, conv.UserID, conv.Topic, conv.Status)

	// 2. Verify the sender is an admin
	adminUserID, _ := repository.GetUserIDByChatID(adminChatID)
	if adminUserID == 0 || !repository.IsUserAdmin(adminUserID) {
		log.Printf("[CHAT-BRIDGE] ❌ Chat %d → user %d is not a linked admin — ignoring reply", adminChatID, adminUserID)
		return
	}

	// 3. Check claim protection: only the claiming admin can reply
	if conv.ClaimedBy != 0 && conv.ClaimedBy != adminUserID {
		claimerName := repository.GetUserDisplayName(conv.ClaimedBy)
		log.Printf("[CHAT-BRIDGE] ⚠️ Admin %d tried to reply to conv #%d claimed by %d (%s)", adminUserID, convID, conv.ClaimedBy, claimerName)
		telegram.SendMessageHTML(adminChatID,
			fmt.Sprintf("⚠️ <b>Этот тикет уже обрабатывает %s.</b>\nВаши ответы не будут отправлены клиенту.", claimerName))
		return
	}

	// 4. Insert admin message (masked name for client anonymity)
	_, err = repository.InsertChatMessage(convID, "admin", "Support Specialist", text, 0)
	if err != nil {
		log.Printf("[CHAT-BRIDGE] ❌ Failed to insert admin reply for conv %d: %v", convID, err)
		return
	}

	log.Printf("[CHAT-BRIDGE] ✅ Admin reply saved to conversation #%d (admin user=%d)", convID, adminUserID)

	// 4. Notify the user via Telegram that they have a new reply
	userChatID := repository.GetUserTelegramChatID(conv.UserID)
	if userChatID != 0 {
		telegram.SendMessageHTML(userChatID,
			fmt.Sprintf("💬 <b>Новое сообщение от поддержки</b>\n\n%s\n\n"+
				"<a href=\"https://xplr.pro/support\">Открыть чат</a>", text))
	} else {
		log.Printf("[CHAT-BRIDGE] User %d has no linked Telegram — skipping TG notification", conv.UserID)
	}
}
