package captcha

import (
	"testing"
	"time"
)

func TestChallengeLimiterPerIP(t *testing.T) {
	l := newChallengeLimiter()
	for i := 0; i < challengePerIPLimit; i++ {
		ok, retry := l.Allow("203.0.113.1")
		if !ok || retry != 0 {
			t.Fatalf("attempt %d: ok=%v retry=%v", i, ok, retry)
		}
	}
	ok, retry := l.Allow("203.0.113.1")
	if ok {
		t.Fatal("expected challenge rate limit")
	}
	if retry <= 0 {
		t.Fatalf("retry = %v", retry)
	}
}

func TestChallengeLimiterWindowExpires(t *testing.T) {
	l := newChallengeLimiter()
	key := "challenge:ip:10.0.0.5"
	l.mu.Lock()
	l.windows[key] = &limitWindow{count: challengePerIPLimit, resetAt: time.Now().Add(-time.Second)}
	l.mu.Unlock()

	ok, retry := l.Allow("10.0.0.5")
	if !ok {
		t.Fatalf("expected expired window, retry=%v", retry)
	}
}
