package service

import (
	"log"

	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/telegram"
)

// NotifyUser sends a notification to a user based on their notification_pref setting.
// pref = 'email' → email only, 'telegram' → TG only, 'both' → both channels.
// subject is used for email; htmlMsg is used for both email body and TG message.
//
// CRITICAL DESIGN: Each channel (Email, Telegram) is fully isolated in its own goroutine
// with its own DB lookup. A failure in one channel NEVER blocks or prevents the other.
// No HTTP context is used — safe for background execution.
func NotifyUser(userID int, subject string, htmlMsg string) {
	pref := repository.GetNotificationPref(userID)
	log.Printf("[NOTIFY-START] UserID: %d, Subject: %q, Pref: %q", userID, subject, pref)

	sendEmail := pref == "email" || pref == "both"
	sendTG := pref == "telegram" || pref == "both"

	// Fallback: if pref is empty or unexpected, send to both channels
	if pref == "" || (!sendEmail && !sendTG) {
		log.Printf("[NOTIFY] ⚠️ User %d has unexpected pref=%q — falling back to 'both'", userID, pref)
		sendEmail = true
		sendTG = true
	}

	// ── Email channel — fully independent goroutine ──
	if sendEmail {
		go func(uid int, subj, body string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[NOTIFY-PANIC] Email goroutine panic for user %d: %v", uid, r)
				}
			}()

			user, err := repository.GetUserByID(uid)
			if err != nil {
				log.Printf("[NOTIFY-FAIL] Email: cannot fetch user %d from DB: %v", uid, err)
				return
			}
			if user.Email == "" {
				log.Printf("[NOTIFY-FAIL] Email: user %d has empty email — skipping", uid)
				return
			}

			log.Printf("[NOTIFY] 📧 Sending email to user %d (%s), subject=%q", uid, user.Email, subj)
			if err := SendGenericEmail(user.Email, subj, body); err != nil {
				log.Printf("[NOTIFY-FAIL] Email to user %d (%s) failed: %v", uid, user.Email, err)
			} else {
				log.Printf("[NOTIFY-SUCCESS] Sent to user %d (%s) via EMAIL", uid, user.Email)
			}
		}(userID, subject, htmlMsg)
	}

	// ── Telegram channel — fully independent goroutine ──
	if sendTG {
		go func(uid int, body string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[NOTIFY-PANIC] Telegram goroutine panic for user %d: %v", uid, r)
				}
			}()

			user, err := repository.GetUserByID(uid)
			if err != nil {
				log.Printf("[NOTIFY-FAIL] Telegram: cannot fetch user %d from DB: %v", uid, err)
				return
			}

			tgValid := user.TelegramChatID.Valid
			tgID := user.TelegramChatID.Int64

			if !tgValid || tgID == 0 {
				log.Printf("[NOTIFY-FAIL] Telegram: user %d has no linked TG (valid=%v, id=%d) — skipping", uid, tgValid, tgID)
				return
			}

			log.Printf("[NOTIFY] 📱 Sending TG to user %d (chat_id=%d)", uid, tgID)
			if err := telegram.SendMessageHTMLSafe(tgID, body); err != nil {
				log.Printf("[NOTIFY-FAIL] Telegram to user %d (chat_id=%d) failed: %v", uid, tgID, err)
			} else {
				log.Printf("[NOTIFY-SUCCESS] Sent to user %d (chat_id=%d) via TELEGRAM", uid, tgID)
			}
		}(userID, htmlMsg)
	}

	log.Printf("[NOTIFY-END] Dispatched notifications for user %d (email=%v, tg=%v)", userID, sendEmail, sendTG)
}

// NotifyUserNews sends a news notification with image-first layout.
// tgCaption is used for Telegram (sendPhoto caption).
// emailBody is the HTML content for the email (image is prepended automatically).
// imageURL is the direct link to the news image.
func NotifyUserNews(userID int, subject string, tgCaption string, emailBody string, imageURL string) {
	pref := repository.GetNotificationPref(userID)
	log.Printf("[NOTIFY-NEWS] UserID: %d, Subject: %q, Pref: %q, HasImage: %v", userID, subject, pref, imageURL != "")

	sendEmail := pref == "email" || pref == "both"
	sendTG := pref == "telegram" || pref == "both"

	if pref == "" || (!sendEmail && !sendTG) {
		sendEmail = true
		sendTG = true
	}

	// ── Email channel — image first ──
	if sendEmail {
		go func(uid int, subj, body, imgURL string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[NOTIFY-PANIC] Email goroutine panic for user %d: %v", uid, r)
				}
			}()

			user, err := repository.GetUserByID(uid)
			if err != nil || user.Email == "" {
				return
			}

			// Build email body: clickable image first, then text
			fullBody := ""
			if imgURL != "" {
				fullBody += `<div style="text-align:center;margin:0 0 24px;"><a href="https://xplr.pro/news"><img src="` + imgURL + `" alt="" style="max-width:100%;height:auto;border-radius:12px;border:1px solid rgba(255,255,255,0.06);" /></a></div>`
			}
			fullBody += body

			if err := SendGenericEmail(user.Email, subj, fullBody); err != nil {
				log.Printf("[NOTIFY-FAIL] Email to user %d failed: %v", uid, err)
			}
		}(userID, subject, emailBody, imageURL)
	}

	// ── Telegram channel — sendPhoto with caption ──
	if sendTG {
		go func(uid int, caption, imgURL string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[NOTIFY-PANIC] Telegram goroutine panic for user %d: %v", uid, r)
				}
			}()

			user, err := repository.GetUserByID(uid)
			if err != nil {
				return
			}
			if !user.TelegramChatID.Valid || user.TelegramChatID.Int64 == 0 {
				return
			}

			tgID := user.TelegramChatID.Int64
			if err := telegram.SendPhotoWithCaption(tgID, imgURL, caption); err != nil {
				log.Printf("[NOTIFY-FAIL] Telegram photo to user %d (chat_id=%d) failed: %v", uid, tgID, err)
			}
		}(userID, tgCaption, imageURL)
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
