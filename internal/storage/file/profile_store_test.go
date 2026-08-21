package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/storage/file"
)

func TestUsersStoreProfileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := file.NewUsersStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	user, err := store.Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}

	profile := users.Profile{
		LastSportType: "Ride",
		LastEquipmentBySport: map[string][]string{
			"Ride": {"bike-1", "helmet-1"},
			"Run":  {"shoe-1"},
		},
	}
	if err := store.PutProfile(user.ID, profile); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetProfile(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.LastSportType != "Ride" {
		t.Fatalf("sport = %q", got.LastSportType)
	}
	if len(got.LastEquipmentBySport["Ride"]) != 2 || got.LastEquipmentBySport["Ride"][0] != "bike-1" {
		t.Fatalf("ride equipment = %#v", got.LastEquipmentBySport["Ride"])
	}

	path := filepath.Join(data.UserDir(dir, "alice"), "profile.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("profile file: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("expected profile.yaml on disk")
	}

	reloaded, err := file.NewUsersStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	again, err := reloaded.GetProfile(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again.LastSportType != "Ride" {
		t.Fatalf("reloaded sport = %q", again.LastSportType)
	}
	if len(again.LastEquipmentBySport["Run"]) != 1 {
		t.Fatalf("reloaded run equipment = %#v", again.LastEquipmentBySport["Run"])
	}
}

func TestUsersStoreSetLastSportAndEquipmentPersist(t *testing.T) {
	dir := t.TempDir()
	store, err := file.NewUsersStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.Create("bob", "Bob", "bob@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}

	if err := store.SetLastSportType(user.ID, "Run"); err != nil {
		t.Fatal(err)
	}
	if err := store.SetLastEquipmentForSport(user.ID, "Run", []string{"eq-1"}); err != nil {
		t.Fatal(err)
	}
	if err := store.TouchUsedSportType(user.ID, "Run"); err != nil {
		t.Fatal(err)
	}
	if err := store.TouchUsedSportType(user.ID, "Ride"); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetProfile(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.LastSportType != "Run" {
		t.Fatalf("sport = %q", got.LastSportType)
	}
	if len(got.LastEquipmentBySport["Run"]) != 1 || got.LastEquipmentBySport["Run"][0] != "eq-1" {
		t.Fatalf("equipment = %#v", got.LastEquipmentBySport["Run"])
	}
	if len(got.UsedSportTypes) != 2 || got.UsedSportTypes[0] != "Ride" || got.UsedSportTypes[1] != "Run" {
		t.Fatalf("used sports = %#v", got.UsedSportTypes)
	}

	if err := store.PruneUsedSportTypes(user.ID, map[string]struct{}{"Ride": {}}); err != nil {
		t.Fatal(err)
	}
	pruned, err := store.GetProfile(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pruned.UsedSportTypes) != 1 || pruned.UsedSportTypes[0] != "Ride" {
		t.Fatalf("after prune %#v", pruned.UsedSportTypes)
	}

	if err := store.SetLastEquipmentForSport(user.ID, "Run", nil); err != nil {
		t.Fatal(err)
	}
	cleared, err := store.GetProfile(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cleared.LastEquipmentBySport["Run"]; ok {
		t.Fatalf("expected Run equipment cleared, got %#v", cleared.LastEquipmentBySport)
	}
}
