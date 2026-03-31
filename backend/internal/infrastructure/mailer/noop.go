package mailer

import "context"

// Noop — заглушка, когда SMTP не настроен.
type Noop struct{}

func (Noop) SendPlain(_ context.Context, _, _, _ string) error {
	return nil
}
