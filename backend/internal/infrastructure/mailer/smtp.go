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
// - 587 (STARTTLS).
// - 465 (implicit TLS).
type SMTP struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

func (m *SMTP) SendPlain(ctx context.Context, toEmail, subject, body string) error {
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
		return m.sendImplicitTLS(ctx, addr, auth, toEmail, msg)
	}

	// STARTTLS (smtp.SendMail поднимет STARTTLS при поддержке сервером)
	err := smtp.SendMail(addr, auth, m.From, []string{toEmail}, msg)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (m *SMTP) sendImplicitTLS(
	ctx context.Context,
	addr string,
	auth smtp.Auth,
	toEmail string,
	msg []byte,
) error {
	d := tls.Dialer{Config: &tls.Config{ServerName: m.Host}}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return wrapper.Wrap(err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		return wrapper.Wrap(err)
	}
	defer c.Close()

	err = c.Auth(auth)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = c.Mail(m.From)
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = c.Rcpt(toEmail)
	if err != nil {
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

	err = w.Close()
	if err != nil {
		return wrapper.Wrap(err)
	}

	err = c.Quit()
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
