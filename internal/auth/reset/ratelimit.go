package reset

import (
	"sync"
	"time"
)

const (
	forgotPerIPLimit    = 10
	forgotPerEmailLimit = 3
	resetPerIPLimit     = 20
	rateLimitWindow     = 15 * time.Minute
)

// Limiter is an in-memory fixed-window rate limiter for password reset endpoints.
type Limiter struct {
	mu      sync.Mutex
	windows map[string]*limitWindow
}

type limitWindow struct {
	count   int
	resetAt time.Time
}

// NewLimiter creates an empty rate limiter.
func NewLimiter() *Limiter {
	return &Limiter{windows: make(map[string]*limitWindow)}
}

// AllowForgot checks IP and email limits for forgot-password requests.
// On success both counters are incremented. retryAfter is set when limited.
func (l *Limiter) AllowForgot(ip, email string) (ok bool, retryAfter time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.pruneLocked(now)

	ipKey := "forgot:ip:" + ip
	emailKey := "forgot:email:" + email

	if retry, limited := l.checkLocked(ipKey, forgotPerIPLimit, now); limited {
		return false, retry
	}
	if retry, limited := l.checkLocked(emailKey, forgotPerEmailLimit, now); limited {
		return false, retry
	}
	l.hitLocked(ipKey, now)
	l.hitLocked(emailKey, now)
	return true, 0
}

// AllowReset checks the IP limit for confirm-reset requests.
func (l *Limiter) AllowReset(ip string) (ok bool, retryAfter time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.pruneLocked(now)

	key := "reset:ip:" + ip
	if retry, limited := l.checkLocked(key, resetPerIPLimit, now); limited {
		return false, retry
	}
	l.hitLocked(key, now)
	return true, 0
}

func (l *Limiter) checkLocked(key string, limit int, now time.Time) (retryAfter time.Duration, limited bool) {
	w := l.windows[key]
	if w == nil || !now.Before(w.resetAt) {
		return 0, false
	}
	if w.count >= limit {
		return time.Until(w.resetAt).Truncate(time.Second), true
	}
	return 0, false
}

func (l *Limiter) hitLocked(key string, now time.Time) {
	w := l.windows[key]
	if w == nil || !now.Before(w.resetAt) {
		l.windows[key] = &limitWindow{count: 1, resetAt: now.Add(rateLimitWindow)}
		return
	}
	w.count++
}

func (l *Limiter) pruneLocked(now time.Time) {
	for k, w := range l.windows {
		if !now.Before(w.resetAt) {
			delete(l.windows, k)
		}
	}
}
