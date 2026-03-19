package service

import (
	"log"

	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/telegram"
)

// NotifyUser sends a notification to a user based on their notification_pref setting.
// pref = 'email' → email only, 'telegram' → TG only, 'both' → both channels.
// subject is used for email; htmlMsg is used for both email body and TG message.
func NotifyUser(userID int, subject string, htmlMsg string) {
	pref := repository.GetNotificationPref(userID)
	log.Printf("[NOTIFY] >>> NotifyUser(userID=%d, subject=%q, pref=%q)", userID, subject, pref)

	user, err := repository.GetUserByID(userID)
	if err != nil {
		log.Printf("[NOTIFY] ❌ Failed to get user %d: %v", userID, err)
		return
	}

	tgValid := user.TelegramChatID.Valid
	tgID := user.TelegramChatID.Int64
	log.Printf("[NOTIFY] User %d: email=%q, telegram_chat_id=%d (valid=%v)", userID, user.Email, tgID, tgValid)

	// Email channel
	if pref == "email" || pref == "both" {
		if user.Email != "" {
			log.Printf("[NOTIFY] 📧 Sending email to user %d (%s), subject=%q", userID, user.Email, subject)
			go func(email, subj, body string) {
				if err := SendGenericEmail(email, subj, body); err != nil {
					log.Printf("[NOTIFY] ❌ Email to user %d (%s) failed: %v", userID, email, err)
				} else {
					log.Printf("[NOTIFY] ✅ Email sent to user %d (%s)", userID, email)
				}
			}(user.Email, subject, htmlMsg)
		} else {
			log.Printf("[NOTIFY] ⚠️ User %d has pref=%q but email is empty — skipping email", userID, pref)
		}
	}

	// Telegram channel
	if pref == "telegram" || pref == "both" {
		if tgValid && tgID != 0 {
			log.Printf("[NOTIFY] 📱 Sending TG to user %d (chat_id=%d)", userID, tgID)
			go func(chatID int64) {
				telegram.SendMessageHTML(chatID, htmlMsg)
				log.Printf("[NOTIFY] ✅ TG sent to user %d (chat_id=%d)", userID, chatID)
			}(tgID)
		} else {
			log.Printf("[NOTIFY] ⚠️ User %d has pref=%q but telegram_chat_id is empty/invalid (valid=%v, id=%d) — skipping TG", userID, pref, tgValid, tgID)
		}
	}

	log.Printf("[NOTIFY] <<< NotifyUser complete for user %d", userID)
}

// NotifyAdmins sends a notification to all users with is_admin=true in the database.
// Uses existing telegram.NotifyAdmins for TG, and sends email to admin emails.
func NotifyAdmins(subject string, htmlMsg string) {
	// TG notification via existing system
	telegram.NotifyAdmins(htmlMsg, "", "")

	// Email notification to all admins
	go func() {
		admins, err := repository.GetAdminEmails()
		if err != nil {
			log.Printf("[NOTIFY] Failed to get admin emails: %v", err)
			return
		}
		for _, email := range admins {
			if err := SendGenericEmail(email, subject, htmlMsg); err != nil {
				log.Printf("[NOTIFY] Email to admin %s failed: %v", email, err)
			}
		}
	}()
}
