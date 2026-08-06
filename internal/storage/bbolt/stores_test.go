package bbolt_test

import (
	"errors"
	"testing"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/users"
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

func TestUserProfileStore(t *testing.T) {
	b := openTestBackend(t)
	store := b.Users()

	user, err := store.Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}

	empty, err := store.GetProfile(user.ID)
	if err != nil || empty.LastSportType != "" || len(empty.LastEquipmentBySport) != 0 {
		t.Fatalf("empty profile: %#v err=%v", empty, err)
	}

	if err := store.SetLastSportType(user.ID, "Ride"); err != nil {
		t.Fatal(err)
	}
	if err := store.SetLastEquipmentForSport(user.ID, "Run", []string{"eq-1"}); err != nil {
		t.Fatal(err)
	}

	profile, err := store.GetProfile(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.LastSportType != "Ride" {
		t.Fatalf("sport: %q", profile.LastSportType)
	}
	if got := profile.LastEquipmentBySport["Run"]; len(got) != 1 || got[0] != "eq-1" {
		t.Fatalf("equipment: %#v", got)
	}

	if err := store.PutProfile(user.ID, users.Profile{
		LastSportType: "Walk",
		LastEquipmentBySport: map[string][]string{
			"Walk": {"eq-2"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	profile, err = store.GetProfile(user.ID)
	if err != nil || profile.LastSportType != "Walk" {
		t.Fatalf("after put: %#v err=%v", profile, err)
	}
}

func TestSocialStoreCreateListAndStatus(t *testing.T) {
	b := openTestBackend(t)
	usersStore := b.Users()
	alice, err := usersStore.Create("alice", "Alice", "alice@example.com", "password123")
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
