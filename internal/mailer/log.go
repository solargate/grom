package mailer

import (
	"context"
	"log/slog"
	"strings"
)

// Log writes messages to slog (for development).
type Log struct{}

func (Log) Send(_ context.Context, msg Message) error {
	if err := validateMessage(msg); err != nil {
		return err
	}
	slog.Info("mail_send",
		"to", strings.Join(msg.To, ","),
		"subject", msg.Subject,
		"text", msg.Text,
		"html", msg.HTML,
	)
	return nil
}
