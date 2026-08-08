package captcha

import (
	"sync"
	"time"
)

const (
	challengePerIPLimit = 60
	challengeWindow     = 15 * time.Minute
)

// challengeLimiter is an in-memory fixed-window limiter for challenge issuance.
type challengeLimiter struct {
	mu      sync.Mutex
	windows map[string]*limitWindow
}

type limitWindow struct {
	count   int
	resetAt time.Time
}

func newChallengeLimiter() *challengeLimiter {
	return &challengeLimiter{windows: make(map[string]*limitWindow)}
}

func (l *challengeLimiter) Allow(ip string) (ok bool, retryAfter time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.pruneLocked(now)

	key := "challenge:ip:" + ip
	w := l.windows[key]
	if w != nil && now.Before(w.resetAt) && w.count >= challengePerIPLimit {
		return false, time.Until(w.resetAt).Truncate(time.Second)
	}
	if w == nil || !now.Before(w.resetAt) {
		l.windows[key] = &limitWindow{count: 1, resetAt: now.Add(challengeWindow)}
		return true, 0
	}
	w.count++
	return true, 0
}

func (l *challengeLimiter) pruneLocked(now time.Time) {
	for k, w := range l.windows {
		if !now.Before(w.resetAt) {
			delete(l.windows, k)
		}
	}
}
