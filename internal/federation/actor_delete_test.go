package federation

import "testing"

func TestIsActorDelete(t *testing.T) {
	actor := "https://remote.test/users/bob"
	if !isActorDelete(actor, actor) {
		t.Fatal("string object should match")
	}
	if !isActorDelete(map[string]any{"id": actor, "type": "Person"}, actor) {
		t.Fatal("Person object should match")
	}
	if !isActorDelete(map[string]any{"id": actor, "type": "Tombstone"}, actor) {
		t.Fatal("Tombstone object should match")
	}
	if isActorDelete("https://remote.test/users/bob/workouts/1", actor) {
		t.Fatal("workout object must not match")
	}
	if isActorDelete(map[string]any{"id": actor + "/workouts/1", "type": "Note"}, actor) {
		t.Fatal("Note must not match")
	}
}
