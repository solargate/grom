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
