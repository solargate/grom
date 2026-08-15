package pat_test

import (
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth/pat"
)

type memRepo struct {
	records []pat.TokenRecord
}

func (m *memRepo) Create(record pat.TokenRecord) error {
	m.records = append(m.records, record)
	return nil
}

func (m *memRepo) ListByUser(userID string) ([]pat.TokenRecord, error) {
	out := make([]pat.TokenRecord, 0)
	for _, r := range m.records {
		if r.UserID == userID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *memRepo) CountByUser(userID string) (int, error) {
	count := 0
	for _, r := range m.records {
		if r.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *memRepo) GetByHash(hash string) (*pat.TokenRecord, error) {
	for _, r := range m.records {
		if r.TokenHash == hash {
			cp := r
			return &cp, nil
		}
	}
	return nil, pat.ErrInvalidToken
}

func (m *memRepo) DeleteByUserAndID(userID, id string) error {
	out := m.records[:0]
	found := false
	for _, r := range m.records {
		if r.UserID == userID && r.ID == id {
			found = true
			continue
		}
		out = append(out, r)
	}
	m.records = out
	if !found {
		return pat.ErrNotFound
	}
	return nil
}

func (m *memRepo) DeleteAllForUser(userID string) error {
	out := m.records[:0]
	for _, r := range m.records {
		if r.UserID == userID {
			continue
		}
		out = append(out, r)
	}
	m.records = out
	return nil
}

func (m *memRepo) setExpiresAt(id string, at time.Time) {
	for i := range m.records {
		if m.records[i].ID == id {
			m.records[i].ExpiresAt = &at
			return
		}
	}
}

func (m *memRepo) countRecords() int {
	return len(m.records)
}

func (m *memRepo) UpdateLastUsed(id string, at time.Time) error {
	for i := range m.records {
		if m.records[i].ID == id {
			m.records[i].LastUsedAt = &at
			return nil
		}
	}
	return pat.ErrNotFound
}

func TestCreateAndAuthenticate(t *testing.T) {
	repo := &memRepo{}
	svc := pat.NewService(repo)

	result, err := svc.Create("user-1", pat.CreateInput{
		Name:   "Script",
		Scopes: []string{pat.ScopeWorkoutsRead},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !strings.HasPrefix(result.Token, pat.TokenPrefix) {
		t.Fatalf("token prefix: %q", result.Token)
	}
	if result.Record.TokenPrefix == "" {
		t.Fatal("expected token_prefix")
	}

	rec, err := svc.Authenticate(result.Token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if rec.UserID != "user-1" {
		t.Fatalf("user id = %q", rec.UserID)
	}

	_, err = svc.Authenticate("grom_pat_invalid")
	if err == nil {
		t.Fatal("expected invalid token")
	}
}

func TestCreateRejectsInvalidScope(t *testing.T) {
	svc := pat.NewService(&memRepo{})
	_, err := svc.Create("user-1", pat.CreateInput{
		Name:   "Script",
		Scopes: []string{"admin:all"},
	})
	if err != pat.ErrInvalidScope {
		t.Fatalf("err = %v", err)
	}
}

func TestMaxTokensPerUser(t *testing.T) {
	repo := &memRepo{}
	svc := pat.NewService(repo)
	for i := 0; i < pat.MaxTokensPerUser; i++ {
		_, err := svc.Create("user-1", pat.CreateInput{
			Name:   "Token",
			Scopes: []string{pat.ScopeWorkoutsRead},
		})
		if err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}
	_, err := svc.Create("user-1", pat.CreateInput{
		Name:   "One too many",
		Scopes: []string{pat.ScopeWorkoutsRead},
	})
	if err != pat.ErrTooManyTokens {
		t.Fatalf("err = %v", err)
	}
}

func TestRevoke(t *testing.T) {
	repo := &memRepo{}
	svc := pat.NewService(repo)
	result, err := svc.Create("user-1", pat.CreateInput{
		Name:   "Script",
		Scopes: []string{pat.ScopeWorkoutsRead},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Revoke("user-1", result.Record.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if _, err := svc.Authenticate(result.Token); err == nil {
		t.Fatal("expected revoked token to fail")
	}
}

func TestAuthenticateExpiredToken(t *testing.T) {
	repo := &memRepo{}
	svc := pat.NewService(repo)
	result, err := svc.Create("user-1", pat.CreateInput{
		Name:   "Script",
		Scopes: []string{pat.ScopeWorkoutsRead},
	})
	if err != nil {
		t.Fatal(err)
	}

	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	repo.setExpiresAt(result.Record.ID, past)

	if _, err := svc.Authenticate(result.Token); err == nil {
		t.Fatal("expected expired token to fail")
	}
	if repo.countRecords() != 0 {
		t.Fatalf("expected expired token to be deleted, records = %d", repo.countRecords())
	}
}

func TestListRemovesExpiredTokens(t *testing.T) {
	repo := &memRepo{}
	svc := pat.NewService(repo)
	result, err := svc.Create("user-1", pat.CreateInput{
		Name:   "Script",
		Scopes: []string{pat.ScopeWorkoutsRead},
	})
	if err != nil {
		t.Fatal(err)
	}

	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	repo.setExpiresAt(result.Record.ID, past)

	list, err := svc.List("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("list = %#v, want empty", list)
	}
	if repo.countRecords() != 0 {
		t.Fatalf("expected expired token to be deleted, records = %d", repo.countRecords())
	}
}

func TestCreateRejectsInvalidName(t *testing.T) {
	svc := pat.NewService(&memRepo{})

	_, err := svc.Create("user-1", pat.CreateInput{
		Name:   "   ",
		Scopes: []string{pat.ScopeWorkoutsRead},
	})
	if err != pat.ErrInvalidRequest {
		t.Fatalf("empty name err = %v", err)
	}

	longName := strings.Repeat("a", pat.MaxNameLen+1)
	_, err = svc.Create("user-1", pat.CreateInput{
		Name:   longName,
		Scopes: []string{pat.ScopeWorkoutsRead},
	})
	if err != pat.ErrInvalidRequest {
		t.Fatalf("long name err = %v", err)
	}
}

func TestCreateRejectsExpirationOverMax(t *testing.T) {
	svc := pat.NewService(&memRepo{})
	_, err := svc.Create("user-1", pat.CreateInput{
		Name:          "Script",
		Scopes:        []string{pat.ScopeWorkoutsRead},
		ExpiresInDays: pat.MaxExpiresDays + 1,
	})
	if err != pat.ErrInvalidRequest {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateRejectsEmptyScopes(t *testing.T) {
	svc := pat.NewService(&memRepo{})
	_, err := svc.Create("user-1", pat.CreateInput{
		Name:   "Script",
		Scopes: []string{},
	})
	if err != pat.ErrInvalidScope {
		t.Fatalf("err = %v", err)
	}
}

func TestRevokeNotFound(t *testing.T) {
	svc := pat.NewService(&memRepo{})
	if err := svc.Revoke("user-1", "missing-id"); err != pat.ErrNotFound {
		t.Fatalf("err = %v", err)
	}
}
