package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// ChatConversation represents a support chat conversation.
type ChatConversation struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Topic     string    `json:"topic"`
	Status    string    `json:"status"` // open, closed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	SenderType     string    `json:"sender_type"` // user, admin
	SenderName     string    `json:"sender_name"`
	Body           string    `json:"body"`
	TgMessageID    int64     `json:"tg_message_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// EnsureChatTables creates chat_conversations and chat_messages if they don't exist.
func EnsureChatTables() error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	_, err := GlobalDB.Exec(`
		CREATE TABLE IF NOT EXISTS chat_conversations (
			id            SERIAL PRIMARY KEY,
			user_id       INTEGER NOT NULL REFERENCES users(id),
			topic         TEXT NOT NULL DEFAULT '',
			status        TEXT NOT NULL DEFAULT 'open',
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_chat_conv_user ON chat_conversations(user_id);
		CREATE INDEX IF NOT EXISTS idx_chat_conv_status ON chat_conversations(status);

		CREATE TABLE IF NOT EXISTS chat_messages (
			id                SERIAL PRIMARY KEY,
			conversation_id   INTEGER NOT NULL REFERENCES chat_conversations(id),
			sender_type       TEXT NOT NULL DEFAULT 'user',
			sender_name       TEXT NOT NULL DEFAULT '',
			body              TEXT NOT NULL DEFAULT '',
			tg_message_id     BIGINT,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_chat_msg_conv ON chat_messages(conversation_id);
	`)
	if err != nil {
		log.Printf("[CHAT] Error creating chat tables: %v", err)
		return err
	}
	log.Println("[CHAT] ✅ Chat tables ensured")
	return nil
}

// CreateConversation opens a new chat conversation for a user.
func CreateConversation(userID int, topic string) (*ChatConversation, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	var conv ChatConversation
	err := GlobalDB.QueryRow(
		`INSERT INTO chat_conversations (user_id, topic) VALUES ($1, $2)
		 RETURNING id, user_id, topic, status, created_at, updated_at`,
		userID, topic,
	).Scan(&conv.ID, &conv.UserID, &conv.Topic, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}
	return &conv, nil
}

// GetOpenConversation returns the user's currently open conversation, if any.
func GetOpenConversation(userID int) (*ChatConversation, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	var conv ChatConversation
	err := GlobalDB.QueryRow(
		`SELECT id, user_id, topic, status, created_at, updated_at
		 FROM chat_conversations WHERE user_id = $1 AND status = 'open'
		 ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&conv.ID, &conv.UserID, &conv.Topic, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetConversationByID returns a conversation by its ID.
func GetConversationByID(convID int) (*ChatConversation, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	var conv ChatConversation
	err := GlobalDB.QueryRow(
		`SELECT id, user_id, topic, status, created_at, updated_at
		 FROM chat_conversations WHERE id = $1`,
		convID,
	).Scan(&conv.ID, &conv.UserID, &conv.Topic, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// CloseConversation marks a conversation as closed.
func CloseConversation(convID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}
	_, err := GlobalDB.Exec(
		`UPDATE chat_conversations SET status = 'closed', updated_at = NOW() WHERE id = $1`,
		convID,
	)
	return err
}

// InsertChatMessage adds a message to a conversation.
func InsertChatMessage(convID int, senderType, senderName, body string, tgMsgID int64) (*ChatMessage, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	var msg ChatMessage
	var tgID sql.NullInt64
	if tgMsgID != 0 {
		tgID = sql.NullInt64{Int64: tgMsgID, Valid: true}
	}
	err := GlobalDB.QueryRow(
		`INSERT INTO chat_messages (conversation_id, sender_type, sender_name, body, tg_message_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, conversation_id, sender_type, sender_name, body, COALESCE(tg_message_id, 0), created_at`,
		convID, senderType, senderName, body, tgID,
	).Scan(&msg.ID, &msg.ConversationID, &msg.SenderType, &msg.SenderName, &msg.Body, &msg.TgMessageID, &msg.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert chat message: %w", err)
	}

	// Touch conversation updated_at
	GlobalDB.Exec(`UPDATE chat_conversations SET updated_at = NOW() WHERE id = $1`, convID)
	return &msg, nil
}

// GetChatMessages returns all messages for a conversation, ordered chronologically.
func GetChatMessages(convID int) ([]ChatMessage, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	rows, err := GlobalDB.Query(
		`SELECT id, conversation_id, sender_type, sender_name, body, COALESCE(tg_message_id, 0), created_at
		 FROM chat_messages WHERE conversation_id = $1 ORDER BY created_at ASC`,
		convID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderType, &m.SenderName, &m.Body, &m.TgMessageID, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// GetConversationByTgMessageID finds the conversation associated with a Telegram message_id.
// Used by the TG bridge to route admin replies.
func GetConversationByTgMessageID(tgMsgID int64) (*ChatConversation, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	var convID int
	err := GlobalDB.QueryRow(
		`SELECT conversation_id FROM chat_messages WHERE tg_message_id = $1 LIMIT 1`,
		tgMsgID,
	).Scan(&convID)
	if err != nil {
		return nil, err
	}
	return GetConversationByID(convID)
}

// UpdateChatMessageTgID sets the tg_message_id on a chat message after it's been forwarded to TG.
func UpdateChatMessageTgID(msgID int, tgMsgID int64) {
	if GlobalDB == nil {
		return
	}
	_, err := GlobalDB.Exec(`UPDATE chat_messages SET tg_message_id = $1 WHERE id = $2`, tgMsgID, msgID)
	if err != nil {
		log.Printf("[CHAT] Failed to update tg_message_id for msg %d: %v", msgID, err)
	}
}

// GetConversationIDByTgReplyMsgID finds the conversation by the tg_message_id that the admin replied to.
func GetConversationIDByTgReplyMsgID(tgMsgID int64) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}
	var convID int
	err := GlobalDB.QueryRow(
		`SELECT conversation_id FROM chat_messages WHERE tg_message_id = $1 LIMIT 1`,
		tgMsgID,
	).Scan(&convID)
	if err != nil {
		return 0, err
	}
	return convID, nil
}

// GetUserTelegramChatID returns the telegram_chat_id for a user.
func GetUserTelegramChatID(userID int) int64 {
	if GlobalDB == nil {
		return 0
	}
	var chatID sql.NullInt64
	_ = GlobalDB.QueryRow(`SELECT telegram_chat_id FROM users WHERE id = $1`, userID).Scan(&chatID)
	if chatID.Valid {
		return chatID.Int64
	}
	return 0
}

// GetUserDisplayName returns the display_name (or email) for a user.
func GetUserDisplayName(userID int) string {
	if GlobalDB == nil {
		return "User"
	}
	var name sql.NullString
	_ = GlobalDB.QueryRow(`SELECT display_name FROM users WHERE id = $1`, userID).Scan(&name)
	if name.Valid && name.String != "" {
		return name.String
	}
	var email string
	_ = GlobalDB.QueryRow(`SELECT email FROM users WHERE id = $1`, userID).Scan(&email)
	if email != "" {
		return email
	}
	return "User"
}

// IsUserAdmin checks if a user has is_admin = true.
func IsUserAdmin(userID int) bool {
	if GlobalDB == nil {
		return false
	}
	var isAdmin bool
	err := GlobalDB.QueryRow(`SELECT COALESCE(is_admin, false) FROM users WHERE id = $1`, userID).Scan(&isAdmin)
	return err == nil && isAdmin
}
