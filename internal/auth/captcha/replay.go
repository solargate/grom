package captcha

import (
	"sync"
	"time"
)

// replayStore remembers used challenge signatures until expiry.
type replayStore struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

func newReplayStore() *replayStore {
	return &replayStore{seen: make(map[string]time.Time)}
}

// Consume marks id as used. Returns false if it was already consumed and not expired.
func (s *replayStore) Consume(id string, expiresAt time.Time) bool {
	if id == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.pruneLocked(now)
	if until, ok := s.seen[id]; ok && now.Before(until) {
		return false
	}
	if expiresAt.IsZero() || !expiresAt.After(now) {
		expiresAt = now.Add(5 * time.Minute)
	}
	s.seen[id] = expiresAt
	return true
}

func (s *replayStore) pruneLocked(now time.Time) {
	for k, until := range s.seen {
		if !now.Before(until) {
			delete(s.seen, k)
		}
	}
}
