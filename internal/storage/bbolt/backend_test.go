package bbolt_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/storage"
	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

func openTestBackend(t *testing.T) *storebbolt.Backend {
	t.Helper()
	dir := t.TempDir()
	backend, err := storebbolt.Open(filepath.Join(dir, "grom.db"), dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = backend.Close() })
	return backend
}

func TestOpenBBoltBackend(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverBBolt,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}
	backend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	if err := backend.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestUsersStoreCreateAndFind(t *testing.T) {
	b := openTestBackend(t)
	created, err := b.Users().Create("Alice", "Alice Name", "Alice@Example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	byEmail, err := b.Users().FindByEmail("alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if byEmail.ID != created.ID || byEmail.PasswordHash == "" {
		t.Fatalf("unexpected user: %#v", byEmail)
	}
	if _, err := b.Users().Create("bob", "Bob", "alice@example.com", "password123"); !errors.Is(err, users.ErrEmailTaken) {
		t.Fatalf("duplicate email: %v", err)
	}
}

func TestWorkoutsStoreCRUDAndListPage(t *testing.T) {
	b := openTestBackend(t)
	repo := b.WorkoutsRepo()

	w1, err := repo.Create("alice", &workouts.Workout{
		Name: "Morning", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	w2, err := repo.Create("alice", &workouts.Workout{
		Name: "Evening", SportType: "Run",
		StartDate: time.Date(2026, 7, 9, 18, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.Get("alice", w1.ID)
	if err != nil || got.Name != "Morning" {
		t.Fatalf("Get: %#v err=%v", got, err)
	}

	page, hasMore, err := repo.ListPage("alice", nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 1 || page[0].ID != w2.ID || !hasMore {
		t.Fatalf("page=%#v hasMore=%v", page, hasMore)
	}
	page2, hasMore2, err := repo.ListPage("alice", &workouts.Cursor{StartDate: page[0].StartDate, ID: page[0].ID}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 1 || page2[0].ID != w1.ID || hasMore2 {
		t.Fatalf("page2=%#v hasMore=%v", page2, hasMore2)
	}

	if err := repo.Delete("alice", w1.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get("alice", w1.ID); !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestWorkoutsBeginCreateWriteMetadata(t *testing.T) {
	b := openTestBackend(t)
	repo := b.WorkoutsRepo()
	w := &workouts.Workout{
		Name: "Track", SportType: "Ride",
		StartDate: time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC),
	}
	created, dirName, cleanup, err := repo.BeginCreate("alice", w)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || dirName == "" {
		t.Fatalf("unexpected begin: %#v %q", created, dirName)
	}
	if err := repo.WriteMetadata("alice", created); err != nil {
		cleanup()
		t.Fatal(err)
	}
	got, err := repo.Get("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Track" {
		t.Fatalf("name=%q", got.Name)
	}
}

func TestFederationFollowersAndInboxMeta(t *testing.T) {
	b := openTestBackend(t)
	followers := b.Federation().Followers()
	if err := followers.Add("alice", federation.InboundFollower{
		ActorURI: "https://remote/users/bob",
		Inbox:    "https://remote/inbox",
		Handle:   "bob@remote",
	}); err != nil {
		t.Fatal(err)
	}
	list, err := followers.List("alice")
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}

	inbox := b.Federation().Inbox()
	w := &workouts.Workout{
		ID: "87654321", Name: "Remote ride", SportType: "Ride",
		StartDate: time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC),
	}
	if err := inbox.Save("alice", "bob@remote", w, nil, nil, map[string]any{"name": "Bob"}); err != nil {
		t.Fatal(err)
	}
	got, err := inbox.Get("alice", "bob", "87654321")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Remote ride" || got.Author.Name != "Bob" {
		t.Fatalf("got=%#v", got)
	}
}
