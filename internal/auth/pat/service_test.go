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
