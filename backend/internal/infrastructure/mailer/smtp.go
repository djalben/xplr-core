package mailer

import (
	"context"
	"fmt"
	"net/smtp"

	"gitlab.com/libs-artifex/wrapper/v2"
)

// SMTP — простая отправка через net/smtp (STARTTLS на порту 587).
type SMTP struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

func (m *SMTP) SendPlain(_ context.Context, toEmail, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	auth := smtp.PlainAuth("", m.User, m.Password, m.Host)

	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	err := smtp.SendMail(addr, auth, m.From, []string{toEmail}, msg)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
