package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/telegram"
	"github.com/gorilla/mux"
)

// ── POST /api/v1/user/chat/start ──
// Creates a new conversation with a topic, or returns the existing open one.
func ChatStartHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Topic string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Topic == "" {
		http.Error(w, "Topic is required", http.StatusBadRequest)
		return
	}

	// Check if user already has an open conversation
	existing, err := repository.GetOpenConversation(userID)
	if err != nil {
		log.Printf("[CHAT] GetOpenConversation error for user %d: %v", userID, err)
	}
	if existing != nil {
		msgs, _ := repository.GetChatMessages(existing.ID)
		if msgs == nil {
			msgs = []repository.ChatMessage{}
		}
		log.Printf("[CHAT] Returning existing conv #%d (topic=%q, msgs=%d) for user %d",
			existing.ID, existing.Topic, len(msgs), userID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"conversation": existing,
			"messages":     msgs,
		})
		return
	}

	// __check__ is a sentinel topic used by the frontend to probe for existing conversations
	if body.Topic == "__check__" {
		log.Printf("[CHAT] No open conversation for user %d (__check__ probe)", userID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"conversation": nil,
			"messages":     []interface{}{},
		})
		return
	}

	conv, err := repository.CreateConversation(userID, body.Topic)
	if err != nil {
		log.Printf("[CHAT] Failed to create conversation for user %d: %v", userID, err)
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	log.Printf("[CHAT] ✅ Conversation #%d created (user=%d, topic=%q)", conv.ID, userID, body.Topic)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conversation": conv,
		"messages":     []interface{}{},
	})
}

// ── GET /api/v1/user/chat/messages/{id} ──
// Returns all messages for a conversation (user must own it).
func ChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	convIDStr := mux.Vars(r)["id"]
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	conv, err := repository.GetConversationByID(convID)
	if err != nil || conv == nil || conv.UserID != userID {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	msgs, err := repository.GetChatMessages(convID)
	if err != nil {
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}
	if msgs == nil {
		msgs = []repository.ChatMessage{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conversation": conv,
		"messages":     msgs,
	})
}

// ── POST /api/v1/user/chat/send/{id} ──
// User sends a message into their conversation. Forwards to Telegram admins.
func ChatSendHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		log.Printf("[CHAT-SEND] ❌ Unauthorized request (no userID in context)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	convIDStr := mux.Vars(r)["id"]
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		log.Printf("[CHAT-SEND] ❌ Invalid conv ID: %q", convIDStr)
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	log.Printf("[CHAT-SEND] >>> User %d sending message to conv %d", userID, convID)

	conv, err := repository.GetConversationByID(convID)
	if err != nil || conv == nil || conv.UserID != userID {
		log.Printf("[CHAT-SEND] ❌ Conv %d not found or not owned by user %d (err=%v, conv=%+v)", convID, userID, err, conv)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	if conv.Status != "open" {
		log.Printf("[CHAT-SEND] ❌ Conv %d is %q, not open", convID, conv.Status)
		http.Error(w, "Conversation is closed", http.StatusBadRequest)
		return
	}

	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Message == "" {
		log.Printf("[CHAT-SEND] ❌ Empty or invalid message body (err=%v)", err)
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	userName := repository.GetUserDisplayName(userID)
	userEmail, _ := repository.GetUserEmail(userID)
	log.Printf("[CHAT-SEND] User %d (%s / %s) sent: %q", userID, userName, userEmail, body.Message)

	msg, err := repository.InsertChatMessage(convID, "user", userName, body.Message, 0)
	if err != nil {
		log.Printf("[CHAT-SEND] ❌ Failed to insert message for conv %d: %v", convID, err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}
	log.Printf("[CHAT-SEND] ✅ Message #%d inserted into conv %d", msg.ID, convID)

	// Forward to Telegram admins SYNCHRONOUSLY.
	// CRITICAL: on Vercel serverless, goroutines launched with `go` are killed
	// as soon as the HTTP response is written. We MUST forward BEFORE responding.
	forwardToTelegramAdmins(conv, msg, userID, userName, userEmail)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

// ── POST /api/v1/user/chat/close/{id} ──
// User closes their conversation.
func ChatCloseHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	convIDStr := mux.Vars(r)["id"]
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	conv, err := repository.GetConversationByID(convID)
	if err != nil || conv == nil || conv.UserID != userID {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	if err := repository.CloseConversation(convID); err != nil {
		http.Error(w, "Failed to close conversation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "closed"})
}

// forwardToTelegramAdmins sends the user's chat message to all admin Telegram accounts.
// The TG message_id is saved so admin replies can be routed back.
// Called SYNCHRONOUSLY — Vercel kills goroutines after HTTP response.
func forwardToTelegramAdmins(conv *repository.ChatConversation, msg *repository.ChatMessage, userID int, userName string, userEmail string) {
	// Panic protection — never let this crash the request
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CHAT-FWD] ❌ PANIC in forwardToTelegramAdmins: %v", r)
		}
	}()

	log.Printf("[CHAT-FWD] [TRACE-1] Start forwarding for message ID: %d (conv=%d, user=%d, email=%s, name=%s)",
		msg.ID, conv.ID, userID, userEmail, userName)

	if telegram.AdminChatIDsProvider == nil {
		log.Printf("[CHAT-FWD] ❌ AdminChatIDsProvider is nil — cannot forward. TELEGRAM_BOT_TOKEN not set?")
		return
	}

	log.Printf("[CHAT-FWD] [TRACE-2] Fetching Admin Chat IDs...")
	ids, err := telegram.AdminChatIDsProvider()
	if err != nil {
		log.Printf("[CHAT-FWD] ❌ AdminChatIDsProvider error: %v", err)
		return
	}
	log.Printf("[CHAT-FWD] [TRACE-3] Found %d admin IDs: %v", len(ids), ids)

	if len(ids) == 0 {
		log.Printf("[CHAT-FWD] ❌ Admin list EMPTY! Check: SELECT email, is_admin, telegram_chat_id FROM users WHERE is_admin = TRUE")
		return
	}

	text := fmt.Sprintf(
		"💬 <b>Чат #%d</b> | <b>%s</b>\n"+
			"📋 Тема: <i>%s</i>\n\n"+
			"%s\n\n"+
			"<i>↩️ Ответьте Reply на это сообщение, чтобы ответить клиенту</i>",
		conv.ID, userName, conv.Topic, msg.Body,
	)

	// Build inline keyboard: Claim button (only if not yet claimed)
	var keyboard *telegram.InlineKeyboardMarkupExported
	if conv.ClaimedBy == 0 {
		keyboard = &telegram.InlineKeyboardMarkupExported{
			InlineKeyboard: [][]telegram.InlineKeyboardButtonExported{
				{
					{Text: "🙋\u200d♂️ Взять в работу", CallbackData: fmt.Sprintf("claim_%d", conv.ID)},
				},
			},
		}
	} else {
		claimerName := repository.GetUserDisplayName(conv.ClaimedBy)
		keyboard = &telegram.InlineKeyboardMarkupExported{
			InlineKeyboard: [][]telegram.InlineKeyboardButtonExported{
				{
					{Text: fmt.Sprintf("✅ В работе у %s", claimerName), CallbackData: "noop"},
				},
			},
		}
	}

	successCount := 0
	for _, chatID := range ids {
		log.Printf("[CHAT-FWD] Sending to admin TG chat %d...", chatID)
		tgMsgID := telegram.SendMessageHTMLWithInlineReturnID(chatID, text, keyboard)
		if tgMsgID != 0 {
			repository.UpdateChatMessageTgID(msg.ID, tgMsgID)
			repository.InsertTgBridge(conv.ID, msg.ID, chatID, tgMsgID)
			successCount++
			log.Printf("[CHAT-FWD] ✅ Sent to admin chat %d → tg_msg_id=%d", chatID, tgMsgID)
		} else {
			log.Printf("[CHAT-FWD] ❌ FAILED to send to admin chat %d (returned 0)", chatID)
		}
	}

	log.Printf("[CHAT-FWD] ✅ Message from [%s] forwarded to [%d/%d] admins via Telegram", userEmail, successCount, len(ids))
}
