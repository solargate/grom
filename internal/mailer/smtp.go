package mailer

import (
	"context"
	"fmt"
	"strings"

	"github.com/solargate/grom/internal/config"
	"github.com/wneessen/go-mail"
)

type smtpMailer struct {
	client *mail.Client
	from   string
}

func newSMTP(cfg config.MailerConfig) (Mailer, error) {
	opts := []mail.Option{
		mail.WithPort(cfg.SMTP.Port),
	}
	user := strings.TrimSpace(cfg.SMTP.Username)
	pass := cfg.SMTP.Password
	if user != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(user),
			mail.WithPassword(pass),
		)
	}

	switch config.MailerEncryption(cfg.SMTP.Encryption) {
	case config.MailerEncryptionTLS:
		opts = append(opts, mail.WithSSL())
	case config.MailerEncryptionNone:
		opts = append(opts, mail.WithTLSPortPolicy(mail.NoTLS))
	default:
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSMandatory))
	}

	client, err := mail.NewClient(cfg.SMTP.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("create smtp mailer: %w", err)
	}
	return &smtpMailer{client: client, from: cfg.From}, nil
}

func (m *smtpMailer) Send(ctx context.Context, msg Message) error {
	if err := validateMessage(msg); err != nil {
		return err
	}

	mmsg := mail.NewMsg()
	if err := mmsg.From(m.from); err != nil {
		return fmt.Errorf("set from: %w", err)
	}
	if err := mmsg.To(msg.To...); err != nil {
		return fmt.Errorf("set to: %w", err)
	}
	mmsg.Subject(msg.Subject)
	mmsg.SetBodyString(mail.TypeTextPlain, msg.Text)
	if html := strings.TrimSpace(msg.HTML); html != "" {
		mmsg.AddAlternativeString(mail.TypeTextHTML, html)
	}

	if err := m.client.DialAndSendWithContext(ctx, mmsg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
