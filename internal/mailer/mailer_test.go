package mailer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/mailer"
)

func TestNew_OffReturnsNop(t *testing.T) {
	m, err := mailer.New(config.MailerConfig{Driver: "off"})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Send(context.Background(), mailer.Message{
		To: []string{"a@b.c"}, Subject: "s", Text: "body",
	}); !errors.Is(err, mailer.ErrDisabled) {
		t.Fatalf("got %v", err)
	}
}

func TestNew_LogValidatesMessage(t *testing.T) {
	m, err := mailer.New(config.MailerConfig{Driver: "log"})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Send(context.Background(), mailer.Message{}); !errors.Is(err, mailer.ErrEmptyMessage) {
		t.Fatalf("empty: %v", err)
	}
	if err := m.Send(context.Background(), mailer.Message{
		To: []string{"a@b.c"}, Subject: "hello", Text: "body",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestNew_UnknownDriver(t *testing.T) {
	_, err := mailer.New(config.MailerConfig{Driver: "ses"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNew_SMTPConstructsMailer(t *testing.T) {
	for _, enc := range []string{"starttls", "tls", "none"} {
		t.Run(enc, func(t *testing.T) {
			cfg := config.MailerConfig{
				Driver: "smtp",
				From:   "noreply@example.com",
			}
			cfg.SMTP.Host = "smtp.example.com"
			cfg.SMTP.Port = 587
			cfg.SMTP.Encryption = enc
			m, err := mailer.New(cfg)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if m == nil {
				t.Fatal("nil mailer")
			}
		})
	}
}

func TestSMTP_SendValidatesBeforeDial(t *testing.T) {
	cfg := config.MailerConfig{
		Driver: "smtp",
		From:   "noreply@example.com",
	}
	cfg.SMTP.Host = "127.0.0.1"
	cfg.SMTP.Port = 1
	cfg.SMTP.Encryption = "none"
	m, err := mailer.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Send(context.Background(), mailer.Message{})
	if !errors.Is(err, mailer.ErrEmptyMessage) {
		t.Fatalf("empty message: %v", err)
	}

	err = m.Send(context.Background(), mailer.Message{
		To:      []string{"user@example.com"},
		Subject: "Reset",
	})
	if !errors.Is(err, mailer.ErrEmptyMessage) {
		t.Fatalf("missing body: %v", err)
	}
}
