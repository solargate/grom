package file

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/solargate/grom/internal/auth/reset"
	"gopkg.in/yaml.v3"
)

const resetTokensFileName = "reset_tokens.yaml"

type resetTokensFile struct {
	Tokens []reset.TokenRecord `yaml:"tokens"`
}

// ResetTokenStore persists password-reset tokens as YAML.
type ResetTokenStore struct {
	path string
	mu   sync.Mutex
}

func NewResetTokenStore(dataDir string) *ResetTokenStore {
	return &ResetTokenStore{path: filepath.Join(dataDir, resetTokensFileName)}
}

func (s *ResetTokenStore) ReplaceForUser(userID string, record reset.TokenRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return err
	}
	out := tokens[:0]
	for _, t := range tokens {
		if t.UserID == userID {
			continue
		}
		out = append(out, t)
	}
	out = append(out, record)
	return s.save(out)
}

func (s *ResetTokenStore) GetByHash(hash string) (*reset.TokenRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	changed := false
	var found *reset.TokenRecord
	out := make([]reset.TokenRecord, 0, len(tokens))
	for i := range tokens {
		t := tokens[i]
		if t.TokenHash != hash {
			out = append(out, t)
			continue
		}
		if !t.ExpiresAt.After(now) {
			changed = true
			continue
		}
		cp := t
		found = &cp
		out = append(out, t)
	}
	if changed {
		if err := s.save(out); err != nil {
			return nil, err
		}
	}
	if found == nil {
		return nil, reset.ErrInvalidToken
	}
	return found, nil
}

func (s *ResetTokenStore) DeleteByHash(hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return err
	}
	out := tokens[:0]
	for _, t := range tokens {
		if t.TokenHash == hash {
			continue
		}
		out = append(out, t)
	}
	return s.save(out)
}

func (s *ResetTokenStore) load() ([]reset.TokenRecord, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var file resetTokensFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return file.Tokens, nil
}

func (s *ResetTokenStore) save(tokens []reset.TokenRecord) error {
	if tokens == nil {
		tokens = []reset.TokenRecord{}
	}
	data, err := yaml.Marshal(resetTokensFile{Tokens: tokens})
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

var _ reset.TokenStore = (*ResetTokenStore)(nil)
