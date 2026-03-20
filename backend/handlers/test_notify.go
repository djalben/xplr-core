package handlers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/repository"
)

// AdminTestNotifyHandler — GET /api/v1/admin/test-notify?userId=ID
// Diagnostic endpoint: directly tests Telegram and SMTP with raw provider responses.
// Returns JSON with full error details for each channel.
func AdminTestNotifyHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		userIDStr = "0"
	}
	userID, _ := strconv.Atoi(userIDStr)

	result := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"user_id":   userID,
	}

	// ── 1. ENV VARS DUMP (masked) ──
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	envCheck := map[string]interface{}{
		"TELEGRAM_BOT_TOKEN": maskEnv(tgToken),
		"SMTP_HOST":          smtpHost,
		"SMTP_PORT":          smtpPort,
		"SMTP_USER":          smtpUser,
		"SMTP_PASS":          maskEnv(smtpPass),
		"SMTP_FROM":          smtpFrom,
		"tg_token_empty":     tgToken == "",
		"smtp_host_empty":    smtpHost == "",
		"smtp_port_empty":    smtpPort == "",
		"smtp_user_empty":    smtpUser == "",
		"smtp_pass_empty":    smtpPass == "",
	}
	result["env_check"] = envCheck

	// ── 2. DB: fetch user data ──
	var userEmail string
	var tgChatID int64
	var tgValid bool
	var notifPref string

	if userID > 0 {
		notifPref = repository.GetNotificationPref(userID)
		user, err := repository.GetUserByID(userID)
		if err != nil {
			result["db_user_error"] = err.Error()
		} else {
			userEmail = user.Email
			tgChatID = user.TelegramChatID.Int64
			tgValid = user.TelegramChatID.Valid
			result["db_user"] = map[string]interface{}{
				"email":            userEmail,
				"telegram_chat_id": tgChatID,
				"tg_valid":         tgValid,
				"notification_pref": notifPref,
			}
		}
	}

	// ── 3. TEST TELEGRAM — direct HTTP call, capture full response ──
	tgResult := testTelegramDirect(tgToken, tgChatID)
	result["telegram_test"] = tgResult

	// ── 4. TEST SMTP — direct connection, capture full response ──
	smtpResult := testSMTPDirect(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom, userEmail)
	result["smtp_test"] = smtpResult

	log.Printf("[TEST-NOTIFY] Full diagnostic result for userId=%d: %+v", userID, result)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}

func maskEnv(val string) string {
	if val == "" {
		return "(EMPTY)"
	}
	if len(val) <= 8 {
		return val[:2] + "***"
	}
	return val[:4] + "***" + val[len(val)-4:]
}

func testTelegramDirect(token string, chatID int64) map[string]interface{} {
	res := map[string]interface{}{}

	if token == "" {
		res["status"] = "SKIP"
		res["error"] = "TELEGRAM_BOT_TOKEN is EMPTY — cannot send"
		return res
	}
	if chatID == 0 {
		res["status"] = "SKIP"
		res["error"] = "telegram_chat_id is 0 — user has no linked Telegram"
		return res
	}

	// Direct API call with full response capture
	testMsg := fmt.Sprintf("🔔 XPLR Notification Test\n\nTimestamp: %s\nChat ID: %d\n\nIf you see this, Telegram channel works.", time.Now().UTC().Format(time.RFC3339), chatID)
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%d&parse_mode=HTML&text=%s",
		token, chatID, url.QueryEscape(testMsg))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		res["status"] = "ERROR"
		res["error"] = fmt.Sprintf("HTTP request failed: %v", err)
		return res
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	res["http_status"] = resp.StatusCode
	res["response_body"] = string(body)

	if resp.StatusCode == 200 {
		res["status"] = "OK"
	} else {
		res["status"] = "FAIL"
		res["error"] = fmt.Sprintf("Telegram API returned %d", resp.StatusCode)
	}

	return res
}

func testSMTPDirect(host, port, user, pass, from, toEmail string) map[string]interface{} {
	res := map[string]interface{}{}

	if host == "" || port == "" {
		res["status"] = "SKIP"
		res["error"] = fmt.Sprintf("SMTP not configured: host=%q, port=%q", host, port)
		return res
	}
	if user == "" || pass == "" {
		res["status"] = "SKIP"
		res["error"] = fmt.Sprintf("SMTP credentials missing: user=%q, pass_len=%d", user, len(pass))
		return res
	}
	if toEmail == "" {
		// Test with the sender address itself
		toEmail = user
		res["note"] = "No user email — sending test to SMTP_USER itself: " + user
	}
	if from == "" {
		from = user
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	subject := "XPLR Notification Test"
	body := fmt.Sprintf("This is a test email from XPLR notification system.\n\nTimestamp: %s\nTo: %s", time.Now().UTC().Format(time.RFC3339), toEmail)

	headers := fmt.Sprintf(
		"From: XPLR <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		from, toEmail, subject,
	)
	msg := []byte(headers + body)

	var sendErr error

	if port == "465" {
		// Implicit TLS
		sendErr = testSMTPImplicitTLS(host, addr, user, pass, from, toEmail, msg)
	} else {
		// STARTTLS (587)
		auth := smtp.PlainAuth("", user, pass, host)
		sendErr = smtp.SendMail(addr, auth, from, []string{toEmail}, msg)
	}

	if sendErr != nil {
		res["status"] = "FAIL"
		res["error"] = sendErr.Error()
	} else {
		res["status"] = "OK"
		res["sent_to"] = toEmail
	}

	return res
}

func testSMTPImplicitTLS(host, addr, user, pass, from, to string, msg []byte) error {
	tlsConfig := &tls.Config{ServerName: host}
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", user, pass, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP AUTH failed (535?): %w", err)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO failed: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}
	if _, err := wc.Write(msg); err != nil {
		return fmt.Errorf("SMTP write body failed: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("SMTP close data failed: %w", err)
	}
	return client.Quit()
}
