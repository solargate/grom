package file_test

import (
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
