// Package httpsig implements ActivityPub cavage-12 HTTP Signatures (RSA-SHA256)
// using code.superseriousbusiness.org/httpsig.
package httpsig

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ssb "code.superseriousbusiness.org/httpsig"
)

const (
	// MaxClockSkew is how far Date may drift from now (Mastodon uses 12h).
	MaxClockSkew = 12 * time.Hour
)

var (
	postHeaders = []string{ssb.RequestTarget, "host", "date", "digest"}
	getHeaders  = []string{ssb.RequestTarget, "host", "date"}
)

// SignPOST sets Date, Digest, and Signature on an outbound POST.
func SignPOST(req *http.Request, body []byte, priv *rsa.PrivateKey, keyID string) error {
	if req == nil || priv == nil || keyID == "" {
		return fmt.Errorf("httpsig: missing request, key, or keyID")
	}
	if req.Header.Get("Date") == "" {
		req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}
	ensureHostHeader(req)
	signer, _, err := ssb.NewSigner(
		[]ssb.Algorithm{ssb.RSA_SHA256},
		ssb.DigestSha256,
		postHeaders,
		ssb.Signature,
		0,
	)
	if err != nil {
		return err
	}
	return signer.SignRequest(priv, keyID, req, body)
}

// SignGET sets Date and Signature on an outbound GET.
func SignGET(req *http.Request, priv *rsa.PrivateKey, keyID string) error {
	if req == nil || priv == nil || keyID == "" {
		return fmt.Errorf("httpsig: missing request, key, or keyID")
	}
	if req.Header.Get("Date") == "" {
		req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}
	ensureHostHeader(req)
	signer, _, err := ssb.NewSigner(
		[]ssb.Algorithm{ssb.RSA_SHA256},
		ssb.DigestSha256,
		getHeaders,
		ssb.Signature,
		0,
	)
	if err != nil {
		return err
	}
	return signer.SignRequest(priv, keyID, req, nil)
}

// VerifyResult is the outcome of verifying an inbound signed request.
type VerifyResult struct {
	KeyID string
}

// Verify checks Signature, Digest (when present / required for POST), and Date skew.
// algo is always RSA-SHA256 (cavage-12 / hs2019 placeholder).
func Verify(req *http.Request, body []byte, pub crypto.PublicKey) (*VerifyResult, error) {
	if req == nil {
		return nil, fmt.Errorf("httpsig: nil request")
	}
	if pub == nil {
		return nil, fmt.Errorf("httpsig: nil public key")
	}
	if err := checkDateSkew(req); err != nil {
		return nil, err
	}
	if req.Method == http.MethodPost || len(body) > 0 || req.Header.Get("Digest") != "" {
		if err := verifyDigest(req.Header.Get("Digest"), body); err != nil {
			return nil, err
		}
	}
	verifier, err := ssb.NewVerifier(req)
	if err != nil {
		return nil, fmt.Errorf("httpsig: %w", err)
	}
	if err := verifier.Verify(pub, ssb.RSA_SHA256); err != nil {
		return nil, fmt.Errorf("httpsig: signature: %w", err)
	}
	return &VerifyResult{KeyID: verifier.KeyId()}, nil
}

// KeyID extracts keyId from the Signature header without full verification.
func KeyID(req *http.Request) (string, error) {
	verifier, err := ssb.NewVerifier(req)
	if err != nil {
		return "", err
	}
	return verifier.KeyId(), nil
}

func ensureHostHeader(req *http.Request) {
	if req.Header.Get("Host") != "" {
		return
	}
	if req.Host != "" {
		req.Header.Set("Host", req.Host)
		return
	}
	if req.URL != nil && req.URL.Host != "" {
		req.Header.Set("Host", req.URL.Host)
	}
}

func checkDateSkew(req *http.Request) error {
	dateHdr := req.Header.Get("Date")
	if dateHdr == "" {
		return fmt.Errorf("httpsig: missing Date header")
	}
	parsed, err := http.ParseTime(dateHdr)
	if err != nil {
		return fmt.Errorf("httpsig: invalid Date: %w", err)
	}
	skew := time.Since(parsed.UTC())
	if skew < 0 {
		skew = -skew
	}
	if skew > MaxClockSkew {
		return fmt.Errorf("httpsig: Date outside allowed skew")
	}
	return nil
}

func verifyDigest(digestHeader string, body []byte) error {
	if digestHeader == "" {
		return fmt.Errorf("httpsig: missing Digest header")
	}
	// Accept both SHA-256= (cavage) and sha-256= (RFC3230 / some Mastodon docs).
	parts := strings.SplitN(digestHeader, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("httpsig: malformed Digest")
	}
	algo := strings.ToUpper(strings.TrimSpace(parts[0]))
	if algo != "SHA-256" {
		return fmt.Errorf("httpsig: unsupported Digest algorithm %q", parts[0])
	}
	sum := sha256.Sum256(body)
	want := base64.StdEncoding.EncodeToString(sum[:])
	got := strings.TrimSpace(parts[1])
	if got != want {
		return fmt.Errorf("httpsig: Digest mismatch")
	}
	return nil
}

// ReadBody reads and restores r.Body so handlers can decode JSON afterward.
func ReadBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}
