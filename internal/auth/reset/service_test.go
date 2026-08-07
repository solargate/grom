package reset_test

import (
	"context"
	"errors"
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
