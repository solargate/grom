package reset_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/auth/reset"
	"github.com/solargate/grom/internal/mailer"
	"github.com/solargate/grom/internal/users"
)

type memMailer struct {
	mu   sync.Mutex
	msgs []mailer.Message
	err  error
}

func (m *memMailer) Send(_ context.Context, msg mailer.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.msgs = append(m.msgs, msg)
	return nil
}

func (m *memMailer) last() mailer.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.msgs) == 0 {
		return mailer.Message{}
	}
	return m.msgs[len(m.msgs)-1]
}

type memUsers struct {
	mu    sync.Mutex
	byID  map[string]*users.User
	email map[string]string
}

func newMemUsers(u *users.User) *memUsers {
	m := &memUsers{
		byID:  map[string]*users.User{u.ID: u},
		email: map[string]string{strings.ToLower(u.Email): u.ID},
	}
	return m
}

func (m *memUsers) FindByEmail(email string) (*users.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.email[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return nil, users.ErrUserNotFound
	}
	return m.byID[id], nil
}
func (m *memUsers) FindByID(id string) (*users.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return nil, users.ErrUserNotFound
	}
	return u, nil
}
func (m *memUsers) FindByNickname(string) (*users.User, error) { return nil, users.ErrUserNotFound }
func (m *memUsers) Search(string, string, int) ([]users.User, error) {
	return nil, nil
}
func (m *memUsers) ListAll() ([]users.User, error) { return nil, nil }
func (m *memUsers) Create(string, string, string, string) (*users.User, error) {
	return nil, users.ErrUserNotFound
}
func (m *memUsers) UpdateProfile(string, string) (*users.User, error) {
	return nil, users.ErrUserNotFound
}
func (m *memUsers) UpdatePassword(userID, passwordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[userID]
	if !ok {
		return users.ErrUserNotFound
	}
	u.PasswordHash = passwordHash
	return nil
}
func (m *memUsers) GetProfile(string) (*users.Profile, error) { return &users.Profile{}, nil }
func (m *memUsers) PutProfile(string, users.Profile) error    { return nil }
func (m *memUsers) SetLastSportType(string, string) error     { return nil }
func (m *memUsers) SetLastEquipmentForSport(string, string, []string) error {
	return nil
}
func (m *memUsers) RemoveEquipmentFromLastSets(string, string) error { return nil }
func (m *memUsers) Delete(string) error                              { return users.ErrUserNotFound }

type memTokens struct {
	mu   sync.Mutex
	byHash map[string]reset.TokenRecord
}

func (s *memTokens) ReplaceForUser(userID string, record reset.TokenRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byHash == nil {
		s.byHash = map[string]reset.TokenRecord{}
	}
	for h, r := range s.byHash {
		if r.UserID == userID {
			delete(s.byHash, h)
		}
	}
	s.byHash[record.TokenHash] = record
	return nil
}

func (s *memTokens) GetByHash(hash string) (*reset.TokenRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.byHash[hash]
	if !ok {
		return nil, reset.ErrInvalidToken
	}
	if !r.ExpiresAt.After(time.Now().UTC()) {
		delete(s.byHash, hash)
		return nil, reset.ErrInvalidToken
	}
	cp := r
	return &cp, nil
}

func (s *memTokens) DeleteByHash(hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byHash, hash)
	return nil
}

func (s *memTokens) DeleteAllForUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for h, r := range s.byHash {
		if r.UserID == userID {
			delete(s.byHash, h)
		}
	}
	return nil
}

func TestRequestAndConfirmReset(t *testing.T) {
	hash, err := auth.HashPassword("oldpassword")
	if err != nil {
		t.Fatal(err)
	}
	u := &users.User{ID: "u1", Email: "alice@example.com", PasswordHash: hash}
	usersRepo := newMemUsers(u)
	tokens := &memTokens{}
	mail := &memMailer{}
	svc := reset.NewService(usersRepo, tokens, mail, reset.Config{
		PublicBaseURL: "https://grom.example.com",
		TokenTTL:      time.Hour,
		ServerName:    "Grom",
		Enabled:       true,
	})

	if err := svc.RequestReset(context.Background(), "unknown@example.com"); err != nil {
		t.Fatal(err)
	}
	if mail.last().Subject != "" {
		t.Fatal("expected no mail for unknown email")
	}

	if err := svc.RequestReset(context.Background(), "Alice@Example.com"); err != nil {
		t.Fatal(err)
	}
	msg := mail.last()
	if msg.Subject == "" || msg.Text == "" || msg.HTML == "" {
		t.Fatalf("expected email content, got %#v", msg)
	}
	idx := strings.Index(msg.Text, "token=")
	if idx < 0 {
		t.Fatalf("token missing in text: %s", msg.Text)
	}
	raw := strings.TrimSpace(msg.Text[idx+len("token="):])
	if i := strings.IndexAny(raw, "\n\r "); i >= 0 {
		raw = raw[:i]
	}

	if err := svc.ConfirmReset(context.Background(), raw, "newpassword"); err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword(u.PasswordHash, "newpassword") {
		t.Fatal("password not updated")
	}
	if err := svc.ConfirmReset(context.Background(), raw, "anotherpass"); !errors.Is(err, reset.ErrInvalidToken) {
		t.Fatalf("reuse: %v", err)
	}
}

func TestConfirmResetWeakPassword(t *testing.T) {
	svc := reset.NewService(newMemUsers(&users.User{ID: "u1", Email: "a@b.c"}), &memTokens{}, &memMailer{}, reset.Config{
		PublicBaseURL: "http://localhost",
		Enabled:       true,
	})
	if err := svc.ConfirmReset(context.Background(), "tok", "short"); !errors.Is(err, reset.ErrWeakPassword) {
		t.Fatalf("got %v", err)
	}
}

func TestRequestReset_MailFailureRollsBackToken(t *testing.T) {
	u := &users.User{ID: "u1", Email: "alice@example.com"}
	tokens := &memTokens{}
	mail := &memMailer{err: errors.New("smtp down")}
	svc := reset.NewService(newMemUsers(u), tokens, mail, reset.Config{
		PublicBaseURL: "https://grom.example.com",
		TokenTTL:      time.Hour,
		Enabled:       true,
	})

	err := svc.RequestReset(context.Background(), "alice@example.com")
	if err == nil {
		t.Fatal("expected mailer error")
	}
	tokens.mu.Lock()
	n := len(tokens.byHash)
	tokens.mu.Unlock()
	if n != 0 {
		t.Fatalf("expected token rollback, got %d tokens", n)
	}
}

func TestRequestReset_ReplacesPreviousToken(t *testing.T) {
	hash, err := auth.HashPassword("oldpassword")
	if err != nil {
		t.Fatal(err)
	}
	u := &users.User{ID: "u1", Email: "alice@example.com", PasswordHash: hash}
	tokens := &memTokens{}
	mail := &memMailer{}
	svc := reset.NewService(newMemUsers(u), tokens, mail, reset.Config{
		PublicBaseURL: "https://grom.example.com/",
		TokenTTL:      time.Hour,
		Enabled:       true,
	})

	if err := svc.RequestReset(context.Background(), "alice@example.com"); err != nil {
		t.Fatal(err)
	}
	first := tokenFromMail(t, mail.last())
	if !strings.Contains(mail.last().Text, "https://grom.example.com/reset-password?token=") {
		t.Fatalf("expected trimmed base URL in link: %s", mail.last().Text)
	}

	if err := svc.RequestReset(context.Background(), "alice@example.com"); err != nil {
		t.Fatal(err)
	}
	second := tokenFromMail(t, mail.last())
	if first == second {
		t.Fatal("expected a new token on second request")
	}

	if err := svc.ConfirmReset(context.Background(), first, "newpassword"); !errors.Is(err, reset.ErrInvalidToken) {
		t.Fatalf("old token: %v", err)
	}
	if err := svc.ConfirmReset(context.Background(), second, "newpassword"); err != nil {
		t.Fatal(err)
	}
}

func TestConfirmReset_ExpiredToken(t *testing.T) {
	u := &users.User{ID: "u1", Email: "alice@example.com"}
	tokens := &memTokens{}
	mail := &memMailer{}
	svc := reset.NewService(newMemUsers(u), tokens, mail, reset.Config{
		PublicBaseURL: "https://grom.example.com",
		TokenTTL:      time.Hour,
		Enabled:       true,
	})
	if err := svc.RequestReset(context.Background(), "alice@example.com"); err != nil {
		t.Fatal(err)
	}
	raw := tokenFromMail(t, mail.last())

	tokens.mu.Lock()
	for h, rec := range tokens.byHash {
		rec.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		tokens.byHash[h] = rec
	}
	tokens.mu.Unlock()

	if err := svc.ConfirmReset(context.Background(), raw, "newpassword"); !errors.Is(err, reset.ErrInvalidToken) {
		t.Fatalf("got %v", err)
	}
}

func TestService_NotConfigured(t *testing.T) {
	svc := reset.NewService(newMemUsers(&users.User{ID: "u1", Email: "a@b.c"}), &memTokens{}, &memMailer{}, reset.Config{
		PublicBaseURL: "http://localhost",
		Enabled:       false,
	})
	if err := svc.RequestReset(context.Background(), "a@b.c"); !errors.Is(err, reset.ErrNotConfigured) {
		t.Fatalf("request: %v", err)
	}
	if err := svc.ConfirmReset(context.Background(), "tok", "password12"); !errors.Is(err, reset.ErrNotConfigured) {
		t.Fatalf("confirm: %v", err)
	}
}

func TestRequestReset_EmptyEmail(t *testing.T) {
	mail := &memMailer{}
	svc := reset.NewService(newMemUsers(&users.User{ID: "u1", Email: "a@b.c"}), &memTokens{}, mail, reset.Config{
		PublicBaseURL: "http://localhost",
		Enabled:       true,
	})
	if err := svc.RequestReset(context.Background(), "  "); err != nil {
		t.Fatal(err)
	}
	if mail.last().Subject != "" {
		t.Fatal("expected no mail for empty email")
	}
}

func TestConfirmReset_EmptyToken(t *testing.T) {
	svc := reset.NewService(newMemUsers(&users.User{ID: "u1", Email: "a@b.c"}), &memTokens{}, &memMailer{}, reset.Config{
		PublicBaseURL: "http://localhost",
		Enabled:       true,
	})
	if err := svc.ConfirmReset(context.Background(), "  ", "password12"); !errors.Is(err, reset.ErrInvalidToken) {
		t.Fatalf("got %v", err)
	}
}

func tokenFromMail(t *testing.T, msg mailer.Message) string {
	t.Helper()
	idx := strings.Index(msg.Text, "token=")
	if idx < 0 {
		t.Fatalf("token missing in text: %s", msg.Text)
	}
	raw := strings.TrimSpace(msg.Text[idx+len("token="):])
	if i := strings.IndexAny(raw, "\n\r "); i >= 0 {
		raw = raw[:i]
	}
	if raw == "" {
		t.Fatal("empty token")
	}
	return raw
}

func TestLimiterForgot(t *testing.T) {
	l := reset.NewLimiter()
	for i := 0; i < 3; i++ {
		ok, _ := l.AllowForgot("1.2.3.4", "a@b.c")
		if !ok {
			t.Fatalf("request %d limited", i)
		}
	}
	ok, retry := l.AllowForgot("1.2.3.4", "a@b.c")
	if ok || retry <= 0 {
		t.Fatalf("expected email limit, ok=%v retry=%v", ok, retry)
	}
}

func TestLimiterForgotIP(t *testing.T) {
	l := reset.NewLimiter()
	for i := 0; i < 10; i++ {
		ok, _ := l.AllowForgot("9.9.9.9", fmt.Sprintf("user%d@example.com", i))
		if !ok {
			t.Fatalf("request %d limited", i)
		}
	}
	ok, retry := l.AllowForgot("9.9.9.9", "other@example.com")
	if ok || retry <= 0 {
		t.Fatalf("expected IP limit, ok=%v retry=%v", ok, retry)
	}
	ok, _ = l.AllowForgot("8.8.8.8", "other@example.com")
	if !ok {
		t.Fatal("different IP should be allowed")
	}
}

func TestLimiterReset(t *testing.T) {
	l := reset.NewLimiter()
	for i := 0; i < 20; i++ {
		ok, _ := l.AllowReset("1.2.3.4")
		if !ok {
			t.Fatalf("request %d limited", i)
		}
	}
	ok, retry := l.AllowReset("1.2.3.4")
	if ok || retry <= 0 {
		t.Fatalf("expected reset IP limit, ok=%v retry=%v", ok, retry)
	}
	ok, _ = l.AllowReset("5.5.5.5")
	if !ok {
		t.Fatal("different IP should be allowed")
	}
}
