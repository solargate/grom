package reset

import (
	"fmt"
	"testing"
	"time"
)

func TestLimiterAllowForgotPerIP(t *testing.T) {
	l := NewLimiter()
	for i := 0; i < forgotPerIPLimit; i++ {
		ok, retry := l.AllowForgot("1.2.3.4", fmt.Sprintf("user%d@example.com", i))
		if !ok || retry != 0 {
			t.Fatalf("attempt %d: ok=%v retry=%v", i, ok, retry)
		}
	}
	ok, retry := l.AllowForgot("1.2.3.4", "extra@example.com")
	if ok {
		t.Fatal("expected IP limit")
	}
	if retry <= 0 {
		t.Fatalf("retry = %v", retry)
	}
}

func TestLimiterAllowForgotPerEmail(t *testing.T) {
	l := NewLimiter()
	for i := 0; i < forgotPerEmailLimit; i++ {
		ok, _ := l.AllowForgot("10.0.0.1", "victim@example.com")
		if !ok {
			t.Fatalf("attempt %d blocked", i)
		}
	}
	ok, retry := l.AllowForgot("10.0.0.99", "victim@example.com")
	if ok {
		t.Fatal("expected email limit")
	}
	if retry <= 0 {
		t.Fatalf("retry = %v", retry)
	}
}

func TestLimiterAllowResetPerIP(t *testing.T) {
	l := NewLimiter()
	for i := 0; i < resetPerIPLimit; i++ {
		ok, retry := l.AllowReset("9.9.9.9")
		if !ok || retry != 0 {
			t.Fatalf("attempt %d: ok=%v retry=%v", i, ok, retry)
		}
	}
	ok, retry := l.AllowReset("9.9.9.9")
	if ok {
		t.Fatal("expected reset IP limit")
	}
	if retry <= 0 {
		t.Fatalf("retry = %v", retry)
	}
}

func TestLimiterWindowExpires(t *testing.T) {
	l := NewLimiter()
	key := "forgot:ip:127.0.0.1"
	l.mu.Lock()
	l.windows[key] = &limitWindow{count: forgotPerIPLimit, resetAt: time.Now().Add(-time.Second)}
	l.mu.Unlock()

	ok, retry := l.AllowForgot("127.0.0.1", "x@example.com")
	if !ok {
		t.Fatalf("expected window expired, retry=%v", retry)
	}
}
