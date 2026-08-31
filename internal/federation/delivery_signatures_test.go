package federation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation/httpsig"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/workouts"
)

func TestPostActivityIncludesHTTPSignature(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	dir := t.TempDir()
	blobs := blobfs.NewStore(dir)
	transport := &sigCaptureTransport{}
	delivery, err := NewDelivery(nil, nil, blobs)
	if err != nil {
		t.Fatal(err)
	}
	delivery.SetClient(&http.Client{Transport: transport})

	workout := &workouts.Workout{
		ID:              "38472901",
		Name:            "Morning run",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 3600,
		Distance:        10000,
	}
	if err := delivery.DeliverWorkout("alice", workout, []string{"https://remote.test/inbox"}, nil, nil); err != nil {
		t.Fatalf("DeliverWorkout: %v", err)
	}
	if len(transport.requests) != 1 {
		t.Fatalf("requests = %d", len(transport.requests))
	}
	req := transport.requests[0]
	if req.Header.Get("Signature") == "" {
		t.Fatal("missing Signature header on outbound POST")
	}
	if req.Header.Get("Digest") == "" {
		t.Fatal("missing Digest header on outbound POST")
	}
}

func TestSignOutboundGETUsesInstanceActorKey(t *testing.T) {
	dir := t.TempDir()
	blobs := blobfs.NewStore(dir)
	req, err := http.NewRequest(http.MethodGet, "https://remote.test/users/bob", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/activity+json")
	if err := signOutboundGET(blobs, req); err != nil {
		t.Fatalf("signOutboundGET: %v", err)
	}
	if req.Header.Get("Signature") == "" {
		t.Fatal("missing Signature on GET")
	}
	ak, err := LoadOrCreateInstanceActorKey(blobs)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := httpsig.Verify(req, nil, &ak.Private.PublicKey); err != nil {
		t.Fatalf("Verify GET signature: %v", err)
	}
}

func TestDeliverWorkoutLikeToOwnerInboxSigned(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	transport := &sigCaptureTransport{}
	// No blob store: DeliverWorkoutLike uses default /users/bob/inbox without remote actor fetch.
	delivery := &Delivery{client: &http.Client{Transport: transport}}

	objectID := "https://remote.test/users/bob/workouts/38472901"
	if _, err := delivery.DeliverWorkoutLike("alice", "bob@remote.test", objectID); err != nil {
		t.Fatalf("DeliverWorkoutLike: %v", err)
	}
	if len(transport.requests) != 1 {
		t.Fatalf("requests = %d", len(transport.requests))
	}
	if !hasSuffix(transport.requests[0].URL.String(), "/users/bob/inbox") {
		t.Fatalf("inbox url = %s", transport.requests[0].URL.String())
	}
	// Unsigned when no key store (documented fallback in postActivity).
	if transport.requests[0].Header.Get("Signature") != "" {
		t.Fatal("expected unsigned delivery without blob store")
	}
}

func TestSignOutboundPOSTRoundTrip(t *testing.T) {
	dir := t.TempDir()
	blobs := blobfs.NewStore(dir)
	body, err := json.Marshal(map[string]any{"type": "Like"})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "https://remote.test/inbox", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if err := signOutboundPOST(blobs, "alice", req, body); err != nil {
		t.Fatalf("signOutboundPOST: %v", err)
	}
	if req.Header.Get("Signature") == "" {
		t.Fatal("missing Signature")
	}
	ak, err := LoadOrCreateUserActorKey(blobs, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := httpsig.Verify(req, body, &ak.Private.PublicKey); err != nil {
		t.Fatalf("Verify POST: %v", err)
	}
}

type sigCaptureTransport struct {
	mu       sync.Mutex
	requests []*http.Request
}

func (s *sigCaptureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	s.mu.Lock()
	clone := req.Clone(req.Context())
	if len(body) > 0 {
		clone.Body = io.NopCloser(bytes.NewReader(body))
		clone.ContentLength = int64(len(body))
	}
	s.requests = append(s.requests, clone)
	s.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusAccepted,
		Body:       io.NopCloser(bytes.NewReader(nil)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
