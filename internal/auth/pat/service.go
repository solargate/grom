package pat

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service manages personal access tokens.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

type CreateInput struct {
	Name           string
	Scopes         []string
	ExpiresInDays  int
	NoExpiration   bool
}

type CreateResult struct {
	Token  string
	Record TokenRecord
}

func (s *Service) Create(userID string, in CreateInput) (*CreateResult, error) {
	name := strings.TrimSpace(in.Name)
	if len(name) < MinNameLen || len(name) > MaxNameLen {
		return nil, ErrInvalidRequest
	}
	scopes, err := normalizeScopes(in.Scopes)
	if err != nil {
		return nil, err
	}

	count, err := s.repo.CountByUser(userID)
	if err != nil {
		return nil, err
	}
	if count >= MaxTokensPerUser {
		return nil, ErrTooManyTokens
	}

	var expiresAt *time.Time
	if in.NoExpiration {
		expiresAt = nil
	} else {
		days := in.ExpiresInDays
		if days <= 0 {
			days = DefaultTTLDays
		}
		if days > MaxExpiresDays {
			return nil, ErrInvalidRequest
		}
		at := s.now().UTC().Add(time.Duration(days) * 24 * time.Hour)
		expiresAt = &at
	}

	raw, hash, err := newTokenSecret()
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	record := TokenRecord{
		ID:          uuid.NewString(),
		TokenHash:   hash,
		TokenPrefix: tokenPrefixDisplay(raw),
		UserID:      userID,
		Name:        name,
		Scopes:      scopes,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
	}
	if err := s.repo.Create(record); err != nil {
		return nil, err
	}
	return &CreateResult{Token: raw, Record: record}, nil
}

func (s *Service) List(userID string) ([]TokenRecord, error) {
	records, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	out := make([]TokenRecord, 0, len(records))
	for _, rec := range records {
		if rec.IsExpired(now) {
			_ = s.repo.DeleteByUserAndID(userID, rec.ID)
			continue
		}
		out = append(out, rec)
	}
	return out, nil
}

func (s *Service) Revoke(userID, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidRequest
	}
	if err := s.repo.DeleteByUserAndID(userID, id); err != nil {
		return err
	}
	return nil
}

func (s *Service) Authenticate(rawToken string) (*TokenRecord, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" || !strings.HasPrefix(rawToken, TokenPrefix) {
		return nil, ErrInvalidToken
	}
	hash := hashToken(rawToken)
	rec, err := s.repo.GetByHash(hash)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	if rec.IsExpired(now) {
		_ = s.repo.DeleteByUserAndID(rec.UserID, rec.ID)
		return nil, ErrInvalidToken
	}
	s.touchLastUsed(rec, now)
	return rec, nil
}

func (s *Service) touchLastUsed(rec *TokenRecord, now time.Time) {
	if rec == nil {
		return
	}
	if rec.LastUsedAt != nil && now.Sub(*rec.LastUsedAt) < LastUsedTouchMinGap {
		return
	}
	at := now
	if err := s.repo.UpdateLastUsed(rec.ID, at); err == nil {
		rec.LastUsedAt = &at
	}
}

func normalizeScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return nil, ErrInvalidScope
	}
	allowed := make(map[string]struct{}, len(ValidScopes))
	for _, s := range ValidScopes {
		allowed[s] = struct{}{}
	}
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, raw := range scopes {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if _, ok := allowed[s]; !ok {
			return nil, ErrInvalidScope
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil, ErrInvalidScope
	}
	return out, nil
}

func newTokenSecret() (raw, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate pat secret: %w", err)
	}
	secret := base64.RawURLEncoding.EncodeToString(buf)
	raw = TokenPrefix + secret
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func tokenPrefixDisplay(raw string) string {
	if len(raw) <= TokenPrefixLen {
		return raw
	}
	return raw[:TokenPrefixLen]
}
