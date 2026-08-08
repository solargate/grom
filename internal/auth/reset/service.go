package reset

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/mailer"
	"github.com/solargate/grom/internal/users"
)

const minPasswordLen = 8

// Config holds reset service settings.
type Config struct {
	PublicBaseURL   string
	TokenTTL        time.Duration
	ServerName      string
	Enabled         bool
}

// Service handles password-reset request and confirmation.
type Service struct {
	users  users.Repository
	tokens TokenStore
	mailer mailer.Mailer
	cfg    Config
	limit  *Limiter
}

// NewService wires a password-reset service.
func NewService(usersRepo users.Repository, tokens TokenStore, m mailer.Mailer, cfg Config) *Service {
	if cfg.TokenTTL <= 0 {
		cfg.TokenTTL = time.Hour
	}
	if strings.TrimSpace(cfg.ServerName) == "" {
		cfg.ServerName = "Grom"
	}
	return &Service{
		users:  usersRepo,
		tokens: tokens,
		mailer: m,
		cfg:    cfg,
		limit:  NewLimiter(),
	}
}

// Enabled reports whether reset is configured.
func (s *Service) Enabled() bool {
	return s != nil && s.cfg.Enabled
}

// Limiter exposes the in-memory rate limiter (for handlers).
func (s *Service) Limiter() *Limiter {
	if s == nil {
		return nil
	}
	return s.limit
}

// RequestReset starts a password reset for email. Unknown emails are a no-op.
func (s *Service) RequestReset(ctx context.Context, email string) error {
	if !s.Enabled() {
		return ErrNotConfigured
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}

	user, err := s.users.FindByEmail(email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil
		}
		return err
	}

	raw, hash, err := newToken()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	rec := TokenRecord{
		TokenHash: hash,
		UserID:    user.ID,
		ExpiresAt: now.Add(s.cfg.TokenTTL),
		CreatedAt: now,
	}
	if err := s.tokens.ReplaceForUser(user.ID, rec); err != nil {
		return err
	}

	link := strings.TrimRight(s.cfg.PublicBaseURL, "/") + "/reset-password?token=" + raw
	msg := buildResetMessage(s.cfg.ServerName, user.Email, link, s.cfg.TokenTTL)
	if err := s.mailer.Send(ctx, msg); err != nil {
		_ = s.tokens.DeleteByHash(hash)
		slog.Error("password_reset_email_failed", "user_id", user.ID, "err", err)
		return err
	}
	slog.Info("password_reset_requested", "user_id", user.ID)
	return nil
}

// ConfirmReset sets a new password using a one-time token.
func (s *Service) ConfirmReset(ctx context.Context, rawToken, newPassword string) error {
	if !s.Enabled() {
		return ErrNotConfigured
	}
	_ = ctx
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return ErrInvalidToken
	}
	if len(newPassword) < minPasswordLen {
		return ErrWeakPassword
	}

	hash := hashToken(rawToken)
	rec, err := s.tokens.GetByHash(hash)
	if err != nil {
		return err
	}

	passwordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(rec.UserID, passwordHash); err != nil {
		return err
	}
	if err := s.tokens.DeleteByHash(hash); err != nil {
		slog.Warn("password_reset_token_delete_failed", "user_id", rec.UserID, "err", err)
	}
	slog.Info("password_reset_completed", "user_id", rec.UserID)
	return nil
}

func newToken() (raw, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate reset token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func buildResetMessage(serverName, to, link string, ttl time.Duration) mailer.Message {
	minutes := int(ttl.Round(time.Minute) / time.Minute)
	if minutes < 1 {
		minutes = 1
	}
	subject := fmt.Sprintf("Reset your %s password", serverName)
	text := fmt.Sprintf(
		"Hello,\n\n"+
			"We received a request to reset the password for your %s account (%s).\n\n"+
			"Open this link to choose a new password:\n%s\n\n"+
			"This link expires in %d minutes and can be used only once.\n\n"+
			"If you did not request a password reset, you can ignore this email.\n",
		serverName, to, link, minutes,
	)
	safeLink := html.EscapeString(link)
	safeName := html.EscapeString(serverName)
	safeTo := html.EscapeString(to)
	htmlBody := fmt.Sprintf(
		"<p>Hello,</p>"+
			"<p>We received a request to reset the password for your %s account (%s).</p>"+
			"<p><a href=\"%s\">Reset your password</a></p>"+
			"<p>This link expires in %d minutes and can be used only once.</p>"+
			"<p>If you did not request a password reset, you can ignore this email.</p>",
		safeName, safeTo, safeLink, minutes,
	)
	return mailer.Message{
		To:      []string{to},
		Subject: subject,
		Text:    text,
		HTML:    htmlBody,
	}
}
