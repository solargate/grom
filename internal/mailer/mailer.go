package mailer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/solargate/grom/internal/config"
)

var (
	// ErrDisabled is returned when mailer.driver is off.
	ErrDisabled = errors.New("mailer is disabled")
	// ErrEmptyMessage is returned when required message fields are missing.
	ErrEmptyMessage = errors.New("mail message is incomplete")
)

// Message is a transactional email payload.
type Message struct {
	To      []string
	Subject string
	Text    string
	HTML    string
}

// Mailer delivers outbound email. Implementations must be safe for concurrent use.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

// New builds a Mailer from config.
func New(cfg config.MailerConfig) (Mailer, error) {
	switch config.MailerDriver(strings.ToLower(strings.TrimSpace(cfg.Driver))) {
	case config.MailerDriverOff, "":
		return Nop{}, nil
	case config.MailerDriverLog:
		return Log{}, nil
	case config.MailerDriverSMTP:
		return newSMTP(cfg)
	default:
		return nil, fmt.Errorf("unknown mailer.driver %q", cfg.Driver)
	}
}

func validateMessage(msg Message) error {
	if len(msg.To) == 0 || strings.TrimSpace(msg.To[0]) == "" {
		return fmt.Errorf("%w: to is required", ErrEmptyMessage)
	}
	if strings.TrimSpace(msg.Subject) == "" {
		return fmt.Errorf("%w: subject is required", ErrEmptyMessage)
	}
	if strings.TrimSpace(msg.Text) == "" {
		return fmt.Errorf("%w: text body is required", ErrEmptyMessage)
	}
	return nil
}
