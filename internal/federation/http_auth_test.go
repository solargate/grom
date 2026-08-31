package federation

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/solargate/grom/internal/federation/httpsig"
)

func TestAuthenticateRequestValidPOST(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	owner := "https://remote.test/users/bob"
	keyID := owner + "#main-key"
	body := []byte(`{"type":"Follow"}`)
	req := httptest.NewRequest(http.MethodPost, "https://grom.test/users/alice/inbox", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/activity+json")
	if err := httpsig.SignPOST(req, body, priv, keyID); err != nil {
		t.Fatal(err)
	}

	gotOwner, gotKeyID, err := AuthenticateRequest(req, body, StaticKeyResolver{
		Keys: map[string]StaticKey{
			keyID: {Public: &priv.PublicKey, Owner: owner},
		},
	})
	if err != nil {
		t.Fatalf("AuthenticateRequest: %v", err)
	}
	if gotOwner != owner || gotKeyID != keyID {
		t.Fatalf("owner=%q keyID=%q", gotOwner, gotKeyID)
	}
}

func TestAuthenticateRequestMissingSignature(t *testing.T) {
	body := []byte(`{}`)
	req := httptest.NewRequest(http.MethodPost, "https://grom.test/inbox", bytes.NewReader(body))
	_, _, err := AuthenticateRequest(req, body, StaticKeyResolver{Keys: map[string]StaticKey{}})
	if err == nil {
		t.Fatal("expected error for missing signature")
	}
	if !errorsIsUnauthorized(err) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestAuthenticateRequestBadDigest(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	keyID := "https://remote.test/users/bob#main-key"
	body := []byte(`{"ok":true}`)
	req := httptest.NewRequest(http.MethodPost, "https://grom.test/inbox", bytes.NewReader(body))
	if err := httpsig.SignPOST(req, body, priv, keyID); err != nil {
		t.Fatal(err)
	}
	_, _, err = AuthenticateRequest(req, []byte(`{"ok":false}`), StaticKeyResolver{
		Keys: map[string]StaticKey{keyID: {Public: &priv.PublicKey, Owner: "https://remote.test/users/bob"}},
	})
	if err == nil {
		t.Fatal("expected digest mismatch")
	}
}

func TestAuthenticateActivityActorKeyMismatch(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	keyOwner := "https://remote.test/users/bob"
	keyID := keyOwner + "#main-key"
	body := []byte(`{"type":"Like","actor":"https://remote.test/users/carol"}`)
	req := httptest.NewRequest(http.MethodPost, "https://grom.test/inbox", bytes.NewReader(body))
	if err := httpsig.SignPOST(req, body, priv, keyID); err != nil {
		t.Fatal(err)
	}
	activity := map[string]any{
		"type":  "Like",
		"actor": "https://remote.test/users/carol",
	}
	err = AuthenticateActivity(req, body, activity, StaticKeyResolver{
		Keys: map[string]StaticKey{keyID: {Public: &priv.PublicKey, Owner: keyOwner}},
	})
	if err != ErrActorKeyMismatch {
		t.Fatalf("got %v, want ErrActorKeyMismatch", err)
	}
}

func TestAuthenticateActivityDeleteAllowsGoneKeySameHost(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	owner := "https://remote.test/users/bob"
	keyID := owner + "#main-key"
	body := []byte(`{"type":"Delete","actor":"https://remote.test/users/bob","object":"https://remote.test/users/bob"}`)
	req := httptest.NewRequest(http.MethodPost, "https://grom.test/inbox", bytes.NewReader(body))
	if err := httpsig.SignPOST(req, body, priv, keyID); err != nil {
		t.Fatal(err)
	}
	activity := map[string]any{
		"type":   "Delete",
		"actor":  owner,
		"object": owner,
	}
	err = AuthenticateActivity(req, body, activity, goneKeyResolver{
		owner: owner,
		keyID: keyID,
	})
	if err != nil {
		t.Fatalf("expected nil for actor delete with gone key, got %v", err)
	}
}

func TestAuthenticateRequestResolveFreshRetry(t *testing.T) {
	stalePriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	freshPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	owner := "https://remote.test/users/bob"
	keyID := owner + "#main-key"
	body := []byte(`{"type":"Follow"}`)
	req := httptest.NewRequest(http.MethodPost, "https://grom.test/inbox", bytes.NewReader(body))
	if err := httpsig.SignPOST(req, body, freshPriv, keyID); err != nil {
		t.Fatal(err)
	}

	resolver := &staleThenFreshResolver{
		stalePub: &stalePriv.PublicKey,
		freshPub: &freshPriv.PublicKey,
		owner:    owner,
	}
	_, _, err = AuthenticateRequest(req, body, resolver)
	if err != nil {
		t.Fatalf("AuthenticateRequest: %v", err)
	}
	if resolver.freshCalls != 1 {
		t.Fatalf("freshCalls = %d, want 1", resolver.freshCalls)
	}
}

func errorsIsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorizedFederation)
}

type goneKeyResolver struct {
	owner string
	keyID string
}

func (g goneKeyResolver) Resolve(keyID string) (any, string, error) {
	if keyID == g.keyID {
		return nil, g.owner, errKeyGone
	}
	return nil, "", errKeyGone
}

type staleThenFreshResolver struct {
	stalePub   any
	freshPub   any
	owner      string
	freshCalls int
}

func (r *staleThenFreshResolver) Resolve(string) (any, string, error) {
	return r.stalePub, r.owner, nil
}

func (r *staleThenFreshResolver) ResolveFresh(string) (any, string, error) {
	r.freshCalls++
	return r.freshPub, r.owner, nil
}
