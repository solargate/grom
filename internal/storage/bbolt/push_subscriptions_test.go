package bbolt_test

import (
	"errors"
	"testing"
	"time"

	"github.com/solargate/grom/internal/notifications"
)

func TestPushSubscriptionStoreCRUD(t *testing.T) {
	backend := openTestBackend(t)
	store := backend.PushSubscriptions()
	now := time.Now().UTC()

	sub := notifications.PushSubscription{
		UserID: "u1", InstallationID: "inst-1", Endpoint: "https://example.test/push",
		P256dh: "p256", Auth: "auth", Platform: "android", UpdatedAt: now,
	}
	if err := store.Upsert(sub); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	list, err := store.ListByUser("u1")
	if err != nil || len(list) != 1 || list[0].Endpoint != sub.Endpoint {
		t.Fatalf("list: %#v err=%v", list, err)
	}

	sub.Endpoint = "https://example.test/push2"
	if err := store.Upsert(sub); err != nil {
		t.Fatal(err)
	}
	list, err = store.ListByUser("u1")
	if err != nil || len(list) != 1 || list[0].Endpoint != sub.Endpoint {
		t.Fatalf("replace: %#v err=%v", list, err)
	}

	if err := store.Upsert(notifications.PushSubscription{
		UserID: "u2", InstallationID: "inst-1", Endpoint: "https://example.test/other",
		P256dh: "p", Auth: "a", UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteByUserAndInstallation("u1", "missing"); !errors.Is(err, notifications.ErrNotFound) {
		t.Fatalf("missing delete: %v", err)
	}
	if err := store.DeleteByUserAndInstallation("u1", "inst-1"); err != nil {
		t.Fatal(err)
	}
	list, err = store.ListByUser("u1")
	if err != nil || len(list) != 0 {
		t.Fatalf("after delete: %#v err=%v", list, err)
	}

	_ = store.Upsert(notifications.PushSubscription{
		UserID: "u1", InstallationID: "gone", Endpoint: "https://example.test/gone",
		P256dh: "p", Auth: "a", UpdatedAt: now,
	})
	if err := store.DeleteByEndpoint("https://example.test/gone"); err != nil {
		t.Fatal(err)
	}
	list, err = store.ListByUser("u1")
	if err != nil || len(list) != 0 {
		t.Fatalf("after endpoint delete: %#v err=%v", list, err)
	}

	if err := store.DeleteAllForUser("u2"); err != nil {
		t.Fatal(err)
	}
	list, err = store.ListByUser("u2")
	if err != nil || len(list) != 0 {
		t.Fatalf("delete all: %#v err=%v", list, err)
	}
}
