package captcha

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
)

const (
	algorithm          = "PBKDF2/SHA-256"
	defaultKeyLength   = 32
	counterMin         = 5000
	counterSpan        = 5000
	testCounterMin     = 5
	testCounterSpan    = 20
)

// Config holds captcha service settings.
type Config struct {
	Enabled        bool
	HMACSecret     string
	Cost           int
	Expires        time.Duration
	// LowCostTest uses a small counter range for unit tests (still honors Cost).
	LowCostTest bool
}

// Service issues and verifies ALTCHA v2 challenges.
type Service struct {
	cfg     Config
	derive  altcha.DeriveKeyFunc
	replay  *replayStore
	limiter *challengeLimiter
}

// NewService wires an ALTCHA captcha service.
func NewService(cfg Config) *Service {
	if cfg.Cost <= 0 {
		cfg.Cost = 1000
	}
	if cfg.Expires <= 0 {
		cfg.Expires = 5 * time.Minute
	}
	return &Service{
		cfg:     cfg,
		derive:  altcha.DeriveKeyPBKDF2(),
		replay:  newReplayStore(),
		limiter: newChallengeLimiter(),
	}
}

// Enabled reports whether captcha verification is required.
func (s *Service) Enabled() bool {
	return s != nil && s.cfg.Enabled
}

// CreateChallenge returns a new signed PoW challenge for the given client IP.
func (s *Service) CreateChallenge(clientIP string) (altcha.Challenge, error) {
	if !s.Enabled() {
		return altcha.Challenge{}, ErrDisabled
	}
	if ok, _ := s.limiter.Allow(clientIP); !ok {
		return altcha.Challenge{}, ErrRateLimited
	}

	counterMinVal := counterMin
	counterSpanVal := counterSpan
	if s.cfg.LowCostTest || s.cfg.Cost < 100 {
		counterMinVal = testCounterMin
		counterSpanVal = testCounterSpan
	}
	counter := counterMinVal + rand.IntN(counterSpanVal)
	expires := time.Now().Add(s.cfg.Expires)

	challenge, err := altcha.CreateChallenge(altcha.CreateChallengeOptions{
		Algorithm:           algorithm,
		DeriveKey:           s.derive,
		HMACSignatureSecret: s.cfg.HMACSecret,
		Cost:                s.cfg.Cost,
		KeyLength:           defaultKeyLength,
		Counter:             &counter,
		ExpiresAt:           &expires,
	})
	if err != nil {
		return altcha.Challenge{}, fmt.Errorf("create challenge: %w", err)
	}
	return challenge, nil
}

// Verify checks a base64-encoded (or raw JSON) ALTCHA payload from the client.
// When captcha is disabled, Verify is a no-op.
func (s *Service) Verify(payload string) error {
	if !s.Enabled() {
		return nil
	}
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return ErrMissing
	}

	raw, err := decodePayloadBytes(payload)
	if err != nil {
		return ErrInvalid
	}

	var p altcha.Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return ErrInvalid
	}
	if p.Challenge.Signature == "" {
		return ErrInvalid
	}

	result, err := altcha.VerifySolution(altcha.VerifySolutionOptions{
		Challenge:           p.Challenge,
		Solution:            p.Solution,
		DeriveKey:           s.derive,
		HMACSignatureSecret: s.cfg.HMACSecret,
	})
	if err != nil {
		return ErrInvalid
	}
	if result.Expired {
		return ErrExpired
	}
	if !result.Verified {
		return ErrInvalid
	}

	expiresAt := time.Now().Add(s.cfg.Expires)
	if p.Challenge.Parameters.ExpiresAt > 0 {
		expiresAt = time.Unix(p.Challenge.Parameters.ExpiresAt, 0)
	}
	if !s.replay.Consume(p.Challenge.Signature, expiresAt) {
		return ErrReplay
	}
	return nil
}

func decodePayloadBytes(payload string) ([]byte, error) {
	if decoded, err := base64.StdEncoding.DecodeString(payload); err == nil {
		trimmed := strings.TrimSpace(string(decoded))
		if strings.HasPrefix(trimmed, "{") {
			return decoded, nil
		}
	}
	if strings.HasPrefix(strings.TrimSpace(payload), "{") {
		return []byte(payload), nil
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(payload); err == nil {
		return decoded, nil
	}
	return nil, ErrInvalid
}

// EncodePayload base64-encodes a challenge+solution payload (for tests/helpers).
func EncodePayload(p altcha.Payload) (string, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}
