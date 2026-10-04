package file_test

import (
	"errors"
	"testing"
	"time"

	"github.com/solargate/grom/internal/notifications"
	"github.com/solargate/grom/internal/storage/file"
)

func TestPushSubscriptionStoreUpsertAndPurge(t *testing.T) {
	dir := t.TempDir()
	store := file.NewPushSubscriptionStore(dir)
	sub := notifications.PushSubscription{
		UserID:         "u1",
		InstallationID: "inst-1",
		Endpoint:       "https://example.test/push",
		P256dh:         "p256",
		Auth:           "auth",
		Platform:       "android",
		UpdatedAt:      time.Now().UTC(),
	}
	if err := store.Upsert(sub); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	list, err := store.ListByUser("u1")
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %#v %v", list, err)
	}
	sub.Endpoint = "https://example.test/push2"
	if err := store.Upsert(sub); err != nil {
		t.Fatalf("upsert replace: %v", err)
	}
	list, err = store.ListByUser("u1")
	if err != nil || len(list) != 1 || list[0].Endpoint != sub.Endpoint {
		t.Fatalf("list after replace: %#v %v", list, err)
	}
	if err := store.DeleteAllForUser("u1"); err != nil {
		t.Fatalf("delete all: %v", err)
	}
	list, err = store.ListByUser("u1")
	if err != nil || len(list) != 0 {
		t.Fatalf("expected empty: %#v %v", list, err)
	}
}

func TestPushSubscriptionStoreDeleteByInstallationEndpointAndListAll(t *testing.T) {
	store := file.NewPushSubscriptionStore(t.TempDir())
	now := time.Now().UTC()
	subs := []notifications.PushSubscription{
		{UserID: "u1", InstallationID: "a", Endpoint: "https://e/1", P256dh: "p", Auth: "a", UpdatedAt: now},
		{UserID: "u1", InstallationID: "b", Endpoint: "https://e/2", P256dh: "p", Auth: "a", UpdatedAt: now},
		{UserID: "u2", InstallationID: "a", Endpoint: "https://e/3", P256dh: "p", Auth: "a", UpdatedAt: now},
	}
	for _, sub := range subs {
		if err := store.Upsert(sub); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.DeleteByUserAndInstallation("u1", "missing"); !errors.Is(err, notifications.ErrNotFound) {
		t.Fatalf("missing delete err = %v", err)
	}
	if err := store.DeleteByUserAndInstallation("u1", "a"); err != nil {
		t.Fatal(err)
	}
	list, err := store.ListByUser("u1")
	if err != nil || len(list) != 1 || list[0].InstallationID != "b" {
		t.Fatalf("after install delete: %#v err=%v", list, err)
	}

	if err := store.DeleteByEndpoint("https://e/2"); err != nil {
		t.Fatal(err)
	}
	list, err = store.ListByUser("u1")
	if err != nil || len(list) != 0 {
		t.Fatalf("after endpoint delete: %#v err=%v", list, err)
	}

	all, err := store.ListAll()
	if err != nil || len(all) != 1 || all[0].UserID != "u2" {
		t.Fatalf("list all: %#v err=%v", all, err)
	}

	u2, err := store.ListByUser("u2")
	if err != nil || len(u2) != 1 {
		t.Fatalf("u2 isolation: %#v err=%v", u2, err)
	}
}
