package bbolt_test

import (
	"errors"
	"testing"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/social"
)

func TestEquipmentStoreCRUD(t *testing.T) {
	b := openTestBackend(t)
	store := b.Equipment()

	created, err := store.Create("alice", &equipment.Equipment{
		Type: equipment.TypeShoes,
		Name: "Trail",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create("bob", &equipment.Equipment{
		Type: equipment.TypeBike, Name: "Road", BikeType: "road",
	}); err != nil {
		t.Fatal(err)
	}

	aliceItems, err := store.List("alice")
	if err != nil || len(aliceItems) != 1 || aliceItems[0].ID != created.ID {
		t.Fatalf("alice list: %#v err=%v", aliceItems, err)
	}

	created.Name = "Trail 2"
	updated, err := store.Update("alice", created)
	if err != nil || updated.Name != "Trail 2" {
		t.Fatalf("update: %#v err=%v", updated, err)
	}

	if err := store.SetDistance("alice", created.ID, 1234); err != nil {
		t.Fatal(err)
	}
	got, err := store.FindByID("alice", created.ID)
	if err != nil || got.Distance != 1234 {
		t.Fatalf("distance: %#v err=%v", got, err)
	}

	if err := store.Delete("alice", created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.FindByID("alice", created.ID); !errors.Is(err, equipment.ErrEquipmentNotFound) {
		t.Fatalf("after delete: %v", err)
	}
}

func TestSocialStoreCreateListAndStatus(t *testing.T) {
	b := openTestBackend(t)
	users := b.Users()
	alice, err := users.Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	store := b.Social()

	follow, err := store.Create(social.Follow{
		FollowerID:     alice.ID,
		TargetHandle:   "bob@local",
		TargetNickname: "bob",
		TargetIsLocal:  true,
		Status:         social.StatusPending,
	})
	if err != nil {
		t.Fatal(err)
	}

	list, err := store.ListByFollower(alice.ID)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %#v err=%v", list, err)
	}

	updated, err := store.UpdateStatus(follow.ID, social.StatusActive)
	if err != nil || updated.Status != social.StatusActive {
		t.Fatalf("status: %#v err=%v", updated, err)
	}

	if err := store.Delete(follow.ID); err != nil {
		t.Fatal(err)
	}
}
