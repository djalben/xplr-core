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

	user, err := repository.GetUserByID(userID)
	if err != nil {
		log.Printf("[NOTIFY] Failed to get user %d: %v", userID, err)
		return
	}

	// Email channel
	if pref == "email" || pref == "both" {
		if user.Email != "" {
			go func(email, subj, body string) {
				if err := SendGenericEmail(email, subj, body); err != nil {
					log.Printf("[NOTIFY] Email to user %d failed: %v", userID, err)
				}
			}(user.Email, subject, htmlMsg)
		}
	}

	// Telegram channel
	if pref == "telegram" || pref == "both" {
		if user.TelegramChatID.Valid && user.TelegramChatID.Int64 != 0 {
			go telegram.SendMessageHTML(user.TelegramChatID.Int64, htmlMsg)
		}
	}
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
