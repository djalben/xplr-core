package service

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// SendVerificationEmail — отправляет письмо с ссылкой подтверждения email.
// Настройки SMTP берутся из переменных окружения:
//   - SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM
//   - APP_DOMAIN — домен фронтенда для формирования ссылки верификации
func SendVerificationEmail(toEmail, token string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	domain := os.Getenv("APP_DOMAIN")

	if host == "" || port == "" {
		log.Printf("[EMAIL] SMTP not configured (SMTP_HOST/SMTP_PORT missing). Verification email for %s skipped. Token: %s", toEmail, token)
		return nil // Не блокируем регистрацию если SMTP не настроен
	}

	if from == "" {
		from = user
	}
	if domain == "" {
		domain = "https://xplr-web.vercel.app"
	}

	verifyURL := fmt.Sprintf("%s/verify?token=%s", domain, token)

	subject := "XPLR — Подтверждение email"
	body := fmt.Sprintf(`Здравствуйте!

Для подтверждения вашего email перейдите по ссылке:

%s

Ссылка действительна 24 часа.

Если вы не регистрировались на XPLR — просто проигнорируйте это письмо.

—
XPLR Team`, verifyURL)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, toEmail, subject, body)

	auth := smtp.PlainAuth("", user, pass, host)
	addr := fmt.Sprintf("%s:%s", host, port)

	if err := smtp.SendMail(addr, auth, from, []string{toEmail}, []byte(msg)); err != nil {
		log.Printf("[EMAIL] Failed to send verification email to %s: %v", toEmail, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[EMAIL] Verification email sent to %s", toEmail)
	return nil
}
