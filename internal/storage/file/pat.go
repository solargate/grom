package file

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/solargate/grom/internal/auth/pat"
	"gopkg.in/yaml.v3"
)

const personalAccessTokensFileName = "personal_access_tokens.yaml"

type patFile struct {
	Tokens []pat.TokenRecord `yaml:"tokens"`
}

// PATStore persists personal access tokens as YAML.
type PATStore struct {
	path string
	mu   sync.Mutex
}

func NewPATStore(dataDir string) *PATStore {
	return &PATStore{path: filepath.Join(dataDir, personalAccessTokensFileName)}
}

func (s *PATStore) Create(record pat.TokenRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return err
	}
	tokens = append(tokens, record)
	return s.save(tokens)
}

func (s *PATStore) ListByUser(userID string) ([]pat.TokenRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]pat.TokenRecord, 0)
	for _, t := range tokens {
		if t.UserID == userID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (s *PATStore) CountByUser(userID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, t := range tokens {
		if t.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (s *PATStore) GetByHash(hash string) (*pat.TokenRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return nil, err
	}
	for i := range tokens {
		if tokens[i].TokenHash == hash {
			cp := tokens[i]
			return &cp, nil
		}
	}
	return nil, pat.ErrInvalidToken
}

func (s *PATStore) DeleteByUserAndID(userID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return err
	}
	out := tokens[:0]
	found := false
	for _, t := range tokens {
		if t.UserID == userID && t.ID == id {
			found = true
			continue
		}
		out = append(out, t)
	}
	if !found {
		return pat.ErrNotFound
	}
	return s.save(out)
}

func (s *PATStore) UpdateLastUsed(id string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return err
	}
	changed := false
	for i := range tokens {
		if tokens[i].ID != id {
			continue
		}
		tokens[i].LastUsedAt = &at
		changed = true
		break
	}
	if !changed {
		return pat.ErrNotFound
	}
	return s.save(tokens)
}

// ListAll returns every personal access token (migration).
func (s *PATStore) ListAll() ([]pat.TokenRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]pat.TokenRecord, len(tokens))
	copy(out, tokens)
	return out, nil
}

// Import writes a PAT record as-is (used by storage migration).
func (s *PATStore) Import(record pat.TokenRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens, err := s.load()
	if err != nil {
		return err
	}
	for i := range tokens {
		if tokens[i].ID == record.ID || tokens[i].TokenHash == record.TokenHash {
			tokens[i] = record
			return s.save(tokens)
		}
	}
	tokens = append(tokens, record)
	return s.save(tokens)
}

func (s *PATStore) load() ([]pat.TokenRecord, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var file patFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return file.Tokens, nil
}

func (s *PATStore) save(tokens []pat.TokenRecord) error {
	if tokens == nil {
		tokens = []pat.TokenRecord{}
	}
	data, err := yaml.Marshal(patFile{Tokens: tokens})
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

var _ pat.Repository = (*PATStore)(nil)
