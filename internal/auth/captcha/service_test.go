package captcha_test

import (
	"testing"
	"time"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
	"github.com/solargate/grom/internal/auth/captcha"
)

func testService(enabled bool) *captcha.Service {
	return captcha.NewService(captcha.Config{
		Enabled:     enabled,
		HMACSecret:  "test-hmac-secret-for-captcha!!",
		Cost:        10,
		Expires:     time.Minute,
		LowCostTest: true,
	})
}

func solvePayload(t *testing.T, challenge altcha.Challenge) string {
	t.Helper()
	solution, err := altcha.SolveChallenge(altcha.SolveChallengeOptions{
		Challenge: challenge,
		DeriveKey: altcha.DeriveKeyPBKDF2(),
	})
	if err != nil {
		t.Fatalf("SolveChallenge: %v", err)
	}
	if solution == nil {
		t.Fatal("SolveChallenge returned nil")
	}
	encoded, err := captcha.EncodePayload(altcha.Payload{
		Challenge: challenge,
		Solution:  *solution,
	})
	if err != nil {
		t.Fatalf("EncodePayload: %v", err)
	}
	return encoded
}

func TestVerify_DisabledNoop(t *testing.T) {
	svc := testService(false)
	if err := svc.Verify(""); err != nil {
		t.Fatalf("disabled verify: %v", err)
	}
	if _, err := svc.CreateChallenge("1.2.3.4"); err == nil {
		t.Fatal("expected ErrDisabled")
	}
}

func TestVerify_RoundTrip(t *testing.T) {
	svc := testService(true)
	challenge, err := svc.CreateChallenge("1.2.3.4")
	if err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}
	payload := solvePayload(t, challenge)
	if err := svc.Verify(payload); err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if err := svc.Verify(payload); err != captcha.ErrReplay {
		t.Fatalf("replay = %v, want ErrReplay", err)
	}
}

func TestVerify_MissingAndInvalid(t *testing.T) {
	svc := testService(true)
	if err := svc.Verify(""); err != captcha.ErrMissing {
		t.Fatalf("empty = %v", err)
	}
	if err := svc.Verify("not-valid"); err != captcha.ErrInvalid {
		t.Fatalf("garbage = %v", err)
	}
}

func TestVerify_Expired(t *testing.T) {
	svc := testService(true)
	expired := time.Now().Add(-time.Minute)
	counter := 10
	challenge, err := altcha.CreateChallenge(altcha.CreateChallengeOptions{
		Algorithm:           "PBKDF2/SHA-256",
		DeriveKey:           altcha.DeriveKeyPBKDF2(),
		HMACSignatureSecret: "test-hmac-secret-for-captcha!!",
		Cost:                10,
		KeyLength:           32,
		Counter:             &counter,
		ExpiresAt:           &expired,
	})
	if err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}
	payload, err := captcha.EncodePayload(altcha.Payload{
		Challenge: challenge,
		Solution:  altcha.Solution{Counter: 0, DerivedKey: "unused"},
	})
	if err != nil {
		t.Fatalf("EncodePayload: %v", err)
	}
	if err := svc.Verify(payload); err != captcha.ErrExpired {
		t.Fatalf("expired = %v, want ErrExpired", err)
	}
}

func TestVerify_HMACFallbackSecret(t *testing.T) {
	svc := captcha.NewService(captcha.Config{
		Enabled:     true,
		HMACSecret:  "secret-a",
		Cost:        10,
		Expires:     time.Minute,
		LowCostTest: true,
	})
	challenge, err := svc.CreateChallenge("10.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	payload := solvePayload(t, challenge)

	other := captcha.NewService(captcha.Config{
		Enabled:     true,
		HMACSecret:  "secret-b",
		Cost:        10,
		Expires:     time.Minute,
		LowCostTest: true,
	})
	if err := other.Verify(payload); err != captcha.ErrInvalid {
		t.Fatalf("wrong hmac = %v", err)
	}
}

func TestCreateChallenge_RateLimited(t *testing.T) {
	svc := testService(true)
	for i := 0; i < 60; i++ {
		if _, err := svc.CreateChallenge("9.9.9.9"); err != nil {
			t.Fatalf("challenge %d: %v", i, err)
		}
	}
	if _, err := svc.CreateChallenge("9.9.9.9"); err != captcha.ErrRateLimited {
		t.Fatalf("expected rate limit, got %v", err)
	}
	if _, err := svc.CreateChallenge("8.8.8.8"); err != nil {
		t.Fatalf("other ip: %v", err)
	}
}
