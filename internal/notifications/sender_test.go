package notifications

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type memPushRepo struct {
	mu   sync.Mutex
	subs []PushSubscription
}

func (m *memPushRepo) Upsert(sub PushSubscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.subs {
		if m.subs[i].UserID == sub.UserID && m.subs[i].InstallationID == sub.InstallationID {
			m.subs[i] = sub
			return nil
		}
	}
	m.subs = append(m.subs, sub)
	return nil
}

func (m *memPushRepo) ListByUser(userID string) ([]PushSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []PushSubscription
	for _, sub := range m.subs {
		if sub.UserID == userID {
			out = append(out, sub)
		}
	}
	return out, nil
}

func (m *memPushRepo) DeleteByUserAndInstallation(userID, installationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := m.subs[:0]
	found := false
	for _, sub := range m.subs {
		if sub.UserID == userID && sub.InstallationID == installationID {
			found = true
			continue
		}
		out = append(out, sub)
	}
	m.subs = out
	if !found {
		return ErrNotFound
	}
	return nil
}

func (m *memPushRepo) DeleteAllForUser(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := m.subs[:0]
	for _, sub := range m.subs {
		if sub.UserID != userID {
			out = append(out, sub)
		}
	}
	m.subs = out
	return nil
}

func (m *memPushRepo) DeleteByEndpoint(endpoint string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := m.subs[:0]
	for _, sub := range m.subs {
		if sub.Endpoint != endpoint {
			out = append(out, sub)
		}
	}
	m.subs = out
	return nil
}

func (m *memPushRepo) ListAll() ([]PushSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]PushSubscription, len(m.subs))
	copy(out, m.subs)
	return out, nil
}

func mustPushDeviceKeys(t *testing.T) (p256dh, auth string) {
	t.Helper()
	priv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	p256dh = base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes())
	authBytes := make([]byte, 16)
	if _, err := rand.Read(authBytes); err != nil {
		t.Fatal(err)
	}
	auth = base64.RawURLEncoding.EncodeToString(authBytes)
	return p256dh, auth
}

func newTestSender(t *testing.T, repo Repository, handler http.HandlerFunc) (*Sender, *httptest.Server) {
	t.Helper()
	keys, err := LoadOrCreateVAPID(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	sender := NewSender(NewVAPIDStore(keys), repo, "mailto:test@localhost")
	sender.client = server.Client()
	return sender, server
}

func TestSendToUserSuccessAndMultiDevice(t *testing.T) {
	var hits atomic.Int32
	repo := &memPushRepo{}
	sender, server := newTestSender(t, repo, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusCreated)
	})
	p256dh, auth := mustPushDeviceKeys(t)
	_ = repo.Upsert(PushSubscription{
		UserID: "u1", InstallationID: "a", Endpoint: server.URL + "/a",
		P256dh: p256dh, Auth: auth, UpdatedAt: time.Now().UTC(),
	})
	p256dh2, auth2 := mustPushDeviceKeys(t)
	_ = repo.Upsert(PushSubscription{
		UserID: "u1", InstallationID: "b", Endpoint: server.URL + "/b",
		P256dh: p256dh2, Auth: auth2, UpdatedAt: time.Now().UTC(),
	})

	sender.SendToUser(context.Background(), "u1", Event{
		Type: TypeWorkoutLiked, Slot: "liked:bob:w1", EventID: "e1",
	})
	if hits.Load() != 2 {
		t.Fatalf("hits = %d, want 2", hits.Load())
	}
	list, _ := repo.ListByUser("u1")
	if len(list) != 2 {
		t.Fatalf("subs should remain: %#v", list)
	}
}

func TestSendToUserDeletesGoneEndpoints(t *testing.T) {
	repo := &memPushRepo{}
	var status atomic.Int32
	status.Store(http.StatusGone)
	sender, server := newTestSender(t, repo, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(int(status.Load()))
	})
	p256dh, auth := mustPushDeviceKeys(t)
	endpoint := server.URL + "/dead"
	_ = repo.Upsert(PushSubscription{
		UserID: "u1", InstallationID: "a", Endpoint: endpoint,
		P256dh: p256dh, Auth: auth, UpdatedAt: time.Now().UTC(),
	})

	sender.SendToUser(context.Background(), "u1", Event{Type: TypeUserFollowed, Slot: "followed:1", EventID: "1"})
	list, _ := repo.ListByUser("u1")
	if len(list) != 0 {
		t.Fatalf("expected dead endpoint removed: %#v", list)
	}

	status.Store(http.StatusNotFound)
	_ = repo.Upsert(PushSubscription{
		UserID: "u1", InstallationID: "a", Endpoint: endpoint,
		P256dh: p256dh, Auth: auth, UpdatedAt: time.Now().UTC(),
	})
	sender.SendToUser(context.Background(), "u1", Event{Type: TypeUserFollowed, Slot: "followed:2", EventID: "2"})
	list, _ = repo.ListByUser("u1")
	if len(list) != 0 {
		t.Fatalf("expected 404 endpoint removed: %#v", list)
	}
}

func TestSendToUserKeepsSubOnServerError(t *testing.T) {
	repo := &memPushRepo{}
	sender, server := newTestSender(t, repo, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	p256dh, auth := mustPushDeviceKeys(t)
	_ = repo.Upsert(PushSubscription{
		UserID: "u1", InstallationID: "a", Endpoint: server.URL + "/x",
		P256dh: p256dh, Auth: auth, UpdatedAt: time.Now().UTC(),
	})
	sender.SendToUser(context.Background(), "u1", Event{Type: TypeWorkoutLiked, Slot: "liked:a:b", EventID: "e"})
	list, _ := repo.ListByUser("u1")
	if len(list) != 1 {
		t.Fatalf("expected sub kept on 502: %#v", list)
	}
}

func TestSendToUserNoopWhenEmpty(t *testing.T) {
	repo := &memPushRepo{}
	sender, _ := newTestSender(t, repo, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected push")
	})
	sender.SendToUser(context.Background(), "u1", Event{Type: TypeWorkoutLiked})
	sender.SendToUser(context.Background(), "", Event{Type: TypeWorkoutLiked})
	var nilSender *Sender
	nilSender.SendToUser(context.Background(), "u1", Event{})
}
