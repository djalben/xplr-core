package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"

	"gitlab.com/libs-artifex/wrapper/v2"
)

// SMTP — отправка через net/smtp.
// Поддержка:
// - 587 (STARTTLS)
// - 465 (implicit TLS)
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

	// Implicit TLS (обычно 465)
	if m.Port == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.Host})
		if err != nil {
			return wrapper.Wrap(err)
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, m.Host)
		if err != nil {
			return wrapper.Wrap(err)
		}
		defer c.Close()

		if err = c.Auth(auth); err != nil {
			return wrapper.Wrap(err)
		}
		if err = c.Mail(m.From); err != nil {
			return wrapper.Wrap(err)
		}
		if err = c.Rcpt(toEmail); err != nil {
			return wrapper.Wrap(err)
		}
		w, err := c.Data()
		if err != nil {
			return wrapper.Wrap(err)
		}
		_, err = w.Write(msg)
		if err != nil {
			_ = w.Close()
			return wrapper.Wrap(err)
		}
		if err = w.Close(); err != nil {
			return wrapper.Wrap(err)
		}
		if err = c.Quit(); err != nil {
			return wrapper.Wrap(err)
		}

		return nil
	}

	// STARTTLS (smtp.SendMail поднимет STARTTLS при поддержке сервером)
	err := smtp.SendMail(addr, auth, m.From, []string{toEmail}, msg)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
