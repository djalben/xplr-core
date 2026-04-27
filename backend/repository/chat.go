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
	ClaimedBy int       `json:"claimed_by"`
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
	AttachmentURL  string    `json:"attachment_url,omitempty"`
	AttachmentType string    `json:"attachment_type,omitempty"`
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

		CREATE TABLE IF NOT EXISTS chat_tg_bridge (
			id                SERIAL PRIMARY KEY,
			conversation_id   INTEGER NOT NULL REFERENCES chat_conversations(id),
			chat_message_id   INTEGER NOT NULL REFERENCES chat_messages(id),
			tg_chat_id        BIGINT NOT NULL,
			tg_message_id     BIGINT NOT NULL,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_chat_tg_bridge_tgmsg ON chat_tg_bridge(tg_message_id);

		DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='chat_conversations' AND column_name='claimed_by') THEN
				ALTER TABLE chat_conversations ADD COLUMN claimed_by INTEGER DEFAULT 0;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='chat_messages' AND column_name='attachment_url') THEN
				ALTER TABLE chat_messages ADD COLUMN attachment_url TEXT DEFAULT '';
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='chat_messages' AND column_name='attachment_type') THEN
				ALTER TABLE chat_messages ADD COLUMN attachment_type TEXT DEFAULT '';
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("[CHAT] Error creating chat tables: %v", err)
		return err
	}
	log.Println("[CHAT] ✅ Chat tables ensured")

	// Verify claimed_by column exists — direct check + fallback
	var colExists bool
	verifyErr := GlobalDB.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='chat_conversations' AND column_name='claimed_by')`,
	).Scan(&colExists)
	if verifyErr != nil {
		log.Printf("[CHAT] ⚠️ Could not verify claimed_by column: %v", verifyErr)
	} else if !colExists {
		log.Println("[CHAT] ⚠️ claimed_by column missing after migration — attempting direct ALTER TABLE")
		_, altErr := GlobalDB.Exec(`ALTER TABLE chat_conversations ADD COLUMN IF NOT EXISTS claimed_by INTEGER DEFAULT 0`)
		if altErr != nil {
			log.Printf("[CHAT] ❌ Direct ALTER TABLE for claimed_by failed: %v", altErr)
		} else {
			log.Println("[CHAT] ✅ claimed_by column added via direct ALTER TABLE fallback")
		}
	} else {
		log.Println("[CHAT] ✅ claimed_by column verified present")
	}

	// SECURITY CLEANUP: reset claimed_by for conversations claimed by non-admin users
	res, cleanErr := GlobalDB.Exec(
		`UPDATE chat_conversations SET claimed_by = 0
		 WHERE claimed_by != 0
		   AND claimed_by NOT IN (SELECT id FROM users WHERE is_admin = TRUE)`)
	if cleanErr != nil {
		log.Printf("[CHAT] ⚠️ Security cleanup of claimed_by failed: %v", cleanErr)
	} else {
		affected, _ := res.RowsAffected()
		if affected > 0 {
			log.Printf("[CHAT] 🛡️ SECURITY CLEANUP: Reset claimed_by on %d conversations claimed by non-admins", affected)
		}
	}

	return nil
}

// AdminChatConversationRow is the admin-facing view of a chat conversation.
type AdminChatConversationRow struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	UserEmail    string `json:"user_email"`
	Topic        string `json:"topic"`
	Status       string `json:"status"`
	ClaimedBy    int    `json:"claimed_by"`
	ClaimerEmail string `json:"claimer_email"`
	MessageCount int    `json:"message_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// GetAllChatConversationsForAdmin returns all chat conversations with user/claimer info.
func GetAllChatConversationsForAdmin(statusFilter string) ([]AdminChatConversationRow, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	q := `SELECT
		cc.id, cc.user_id, u.email,
		cc.topic, cc.status, COALESCE(cc.claimed_by, 0),
		COALESCE(adm.email, ''),
		(SELECT COUNT(*) FROM chat_messages cm WHERE cm.conversation_id = cc.id),
		cc.created_at, cc.updated_at
	FROM chat_conversations cc
	JOIN users u ON u.id = cc.user_id
	LEFT JOIN users adm ON adm.id = cc.claimed_by`
	var args []interface{}
	if statusFilter != "" {
		q += ` WHERE cc.status = $1`
		args = append(args, statusFilter)
	}
	q += ` ORDER BY cc.updated_at DESC LIMIT 100`

	rows, err := GlobalDB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []AdminChatConversationRow
	for rows.Next() {
		var c AdminChatConversationRow
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&c.ID, &c.UserID, &c.UserEmail, &c.Topic, &c.Status,
			&c.ClaimedBy, &c.ClaimerEmail, &c.MessageCount,
			&createdAt, &updatedAt); err != nil {
			log.Printf("[CHAT-ADMIN] scan error: %v", err)
			continue
		}
		c.CreatedAt = createdAt.Format(time.RFC3339)
		c.UpdatedAt = updatedAt.Format(time.RFC3339)
		convs = append(convs, c)
	}
	if convs == nil {
		convs = []AdminChatConversationRow{}
	}
	return convs, nil
}

// CreateConversation opens a new chat conversation for a user.
func CreateConversation(userID int, topic string) (*ChatConversation, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	var conv ChatConversation
	err := GlobalDB.QueryRow(
		`INSERT INTO chat_conversations (user_id, topic) VALUES ($1, $2)
		 RETURNING id, user_id, topic, status, COALESCE(claimed_by,0), created_at, updated_at`,
		userID, topic,
	).Scan(&conv.ID, &conv.UserID, &conv.Topic, &conv.Status, &conv.ClaimedBy, &conv.CreatedAt, &conv.UpdatedAt)
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
		`SELECT id, user_id, topic, status, COALESCE(claimed_by,0), created_at, updated_at
		 FROM chat_conversations WHERE user_id = $1 AND status = 'open'
		 ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&conv.ID, &conv.UserID, &conv.Topic, &conv.Status, &conv.ClaimedBy, &conv.CreatedAt, &conv.UpdatedAt)
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
		`SELECT id, user_id, topic, status, COALESCE(claimed_by,0), created_at, updated_at
		 FROM chat_conversations WHERE id = $1`,
		convID,
	).Scan(&conv.ID, &conv.UserID, &conv.Topic, &conv.Status, &conv.ClaimedBy, &conv.CreatedAt, &conv.UpdatedAt)
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
		`UPDATE chat_conversations SET status = 'closed', claimed_by = 0, updated_at = NOW() WHERE id = $1`,
		convID,
	)
	return err
}

// ClaimConversation atomically sets claimed_by if not already claimed. Returns true if claimed successfully.
func ClaimConversation(convID int, adminUserID int) (bool, error) {
	if GlobalDB == nil {
		return false, fmt.Errorf("database connection not initialized")
	}

	// Pre-check: log current state of the conversation
	var curStatus string
	var curClaimedBy sql.NullInt64
	preErr := GlobalDB.QueryRow(
		`SELECT status, claimed_by FROM chat_conversations WHERE id = $1`, convID,
	).Scan(&curStatus, &curClaimedBy)
	if preErr != nil {
		log.Printf("[CHAT-CLAIM-DB] Pre-check failed for conv %d: %v", convID, preErr)
	} else {
		log.Printf("[CHAT-CLAIM-DB] Pre-check conv %d: status=%q, claimed_by=%v (valid=%v)",
			convID, curStatus, curClaimedBy.Int64, curClaimedBy.Valid)
	}

	res, err := GlobalDB.Exec(
		`UPDATE chat_conversations SET claimed_by = $1, updated_at = NOW()
		 WHERE id = $2 AND (claimed_by = 0 OR claimed_by IS NULL)`,
		adminUserID, convID,
	)
	if err != nil {
		log.Printf("[CHAT-CLAIM-DB] UPDATE failed: conv=%d, admin=%d, err=%v", convID, adminUserID, err)
		return false, err
	}
	rows, _ := res.RowsAffected()
	log.Printf("[CHAT-CLAIM-DB] UPDATE result: conv=%d, admin=%d, rowsAffected=%d", convID, adminUserID, rows)
	return rows > 0, nil
}

// GetClaimedOpenConversation returns the open conversation claimed by a given admin user.
// Used as fallback when bridge lookup fails — admin can still reply to their claimed conversation.
func GetClaimedOpenConversation(adminUserID int) (*ChatConversation, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	var conv ChatConversation
	err := GlobalDB.QueryRow(
		`SELECT id, user_id, topic, status, COALESCE(claimed_by,0), created_at, updated_at
		 FROM chat_conversations WHERE claimed_by = $1 AND status = 'open'
		 ORDER BY updated_at DESC LIMIT 1`,
		adminUserID,
	).Scan(&conv.ID, &conv.UserID, &conv.Topic, &conv.Status, &conv.ClaimedBy, &conv.CreatedAt, &conv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// TgBridgeEntry represents a single TG bridge mapping row.
type TgBridgeEntry struct {
	TgChatID    int64
	TgMessageID int64
}

// GetTgBridgeForConversation returns all TG bridge entries for a conversation.
func GetTgBridgeForConversation(convID int) ([]TgBridgeEntry, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	rows, err := GlobalDB.Query(
		`SELECT DISTINCT tg_chat_id, tg_message_id FROM chat_tg_bridge WHERE conversation_id = $1`,
		convID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []TgBridgeEntry
	for rows.Next() {
		var e TgBridgeEntry
		if err := rows.Scan(&e.TgChatID, &e.TgMessageID); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetLatestTgBridgePerAdmin returns the latest TG bridge entry per admin for a conversation.
func GetLatestTgBridgePerAdmin(convID int) ([]TgBridgeEntry, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	rows, err := GlobalDB.Query(
		`SELECT tg_chat_id, MAX(tg_message_id) FROM chat_tg_bridge
		 WHERE conversation_id = $1 GROUP BY tg_chat_id`,
		convID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []TgBridgeEntry
	for rows.Next() {
		var e TgBridgeEntry
		if err := rows.Scan(&e.TgChatID, &e.TgMessageID); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// InsertChatMessage adds a message to a conversation.
func InsertChatMessage(convID int, senderType, senderName, body string, tgMsgID int64) (*ChatMessage, error) {
	return InsertChatMessageWithAttachment(convID, senderType, senderName, body, tgMsgID, "", "")
}

// InsertChatMessageWithAttachment adds a message with optional file attachment.
func InsertChatMessageWithAttachment(convID int, senderType, senderName, body string, tgMsgID int64, attachmentURL, attachmentType string) (*ChatMessage, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	var msg ChatMessage
	var tgID sql.NullInt64
	if tgMsgID != 0 {
		tgID = sql.NullInt64{Int64: tgMsgID, Valid: true}
	}
	err := GlobalDB.QueryRow(
		`INSERT INTO chat_messages (conversation_id, sender_type, sender_name, body, tg_message_id, attachment_url, attachment_type)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, conversation_id, sender_type, sender_name, body, COALESCE(tg_message_id, 0), COALESCE(attachment_url, ''), COALESCE(attachment_type, ''), created_at`,
		convID, senderType, senderName, body, tgID, attachmentURL, attachmentType,
	).Scan(&msg.ID, &msg.ConversationID, &msg.SenderType, &msg.SenderName, &msg.Body, &msg.TgMessageID, &msg.AttachmentURL, &msg.AttachmentType, &msg.CreatedAt)
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
		`SELECT id, conversation_id, sender_type, sender_name, body, COALESCE(tg_message_id, 0),
		        COALESCE(attachment_url, ''), COALESCE(attachment_type, ''), created_at
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
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderType, &m.SenderName, &m.Body, &m.TgMessageID, &m.AttachmentURL, &m.AttachmentType, &m.CreatedAt); err != nil {
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
	} else {
		log.Printf("[CHAT] ✅ Saved tg_message_id=%d for chat msg #%d", tgMsgID, msgID)
	}
}

// InsertTgBridge saves a mapping between a TG message sent to an admin and the chat conversation/message.
// This supports multiple admins each receiving a different TG message_id.
func InsertTgBridge(convID int, chatMsgID int, tgChatID int64, tgMsgID int64) {
	if GlobalDB == nil {
		return
	}
	_, err := GlobalDB.Exec(
		`INSERT INTO chat_tg_bridge (conversation_id, chat_message_id, tg_chat_id, tg_message_id) VALUES ($1, $2, $3, $4)`,
		convID, chatMsgID, tgChatID, tgMsgID,
	)
	if err != nil {
		log.Printf("[CHAT] Failed to insert tg_bridge (conv=%d, tgMsg=%d): %v", convID, tgMsgID, err)
	} else {
		log.Printf("[CHAT] ✅ Bridge saved: conv=%d, chatMsg=%d, tgChat=%d, tgMsg=%d", convID, chatMsgID, tgChatID, tgMsgID)
	}
}

// GetConversationIDByTgReplyMsgID finds the conversation by the tg_message_id that the admin replied to.
// First checks the bridge table (supports multi-admin), then falls back to chat_messages column.
func GetConversationIDByTgReplyMsgID(tgMsgID int64) (int, error) {
	if GlobalDB == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	// Primary: check bridge table
	var convID int
	err := GlobalDB.QueryRow(
		`SELECT conversation_id FROM chat_tg_bridge WHERE tg_message_id = $1 LIMIT 1`,
		tgMsgID,
	).Scan(&convID)
	if err == nil && convID != 0 {
		log.Printf("[CHAT] Bridge lookup: tg_message_id=%d → conv=%d", tgMsgID, convID)
		return convID, nil
	}

	// Fallback: legacy column on chat_messages
	err = GlobalDB.QueryRow(
		`SELECT conversation_id FROM chat_messages WHERE tg_message_id = $1 LIMIT 1`,
		tgMsgID,
	).Scan(&convID)
	if err != nil {
		log.Printf("[CHAT] Bridge lookup FAILED for tg_message_id=%d: %v", tgMsgID, err)
		return 0, err
	}
	log.Printf("[CHAT] Legacy lookup: tg_message_id=%d → conv=%d", tgMsgID, convID)
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

// hardcodedAdminEmails is a safety-net whitelist so these admins always pass IsUserAdmin.
var hardcodedAdminEmails = map[string]bool{
	"aalabin5@gmail.com": true,
	"vardump@inbox.ru":   true,
}

// IsUserAdmin checks if a user has is_admin = true, with a hardcoded email whitelist fallback.
func IsUserAdmin(userID int) bool {
	if GlobalDB == nil {
		return false
	}
	var isAdmin bool
	err := GlobalDB.QueryRow(`SELECT COALESCE(is_admin, false) FROM users WHERE id = $1`, userID).Scan(&isAdmin)
	if err == nil && isAdmin {
		return true
	}
	// Fallback: check hardcoded whitelist by email
	var email string
	_ = GlobalDB.QueryRow(`SELECT email FROM users WHERE id = $1`, userID).Scan(&email)
	if hardcodedAdminEmails[email] {
		log.Printf("[ADMIN] User %d (%s) matched hardcoded admin whitelist", userID, email)
		return true
	}
	return false
}
