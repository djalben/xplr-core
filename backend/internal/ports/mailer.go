package ports

import "context"

// Mailer — отправка писем (верификация, сброс пароля).
type Mailer interface {
	SendPlain(ctx context.Context, toEmail, subject, body string) error
}
