package handlers

import (
	"encoding/json"
	"log"
	"net/http"
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
	Message *tgMessage `json:"message"`
}

type tgMessage struct {
	Chat tgChat `json:"chat"`
	Text string `json:"text"`
}

type tgChat struct {
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

	if update.Message == nil || update.Message.Text == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	text := strings.TrimSpace(update.Message.Text)
	chatID := update.Message.Chat.ID

	log.Printf("[TG-WEBHOOK] Received message: chat_id=%d, text=%q", chatID, text)

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
