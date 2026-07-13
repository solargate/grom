package federation

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeliverWorkoutDelete(t *testing.T) {
	var received []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var activity map[string]any
		if err := json.Unmarshal(body, &activity); err != nil {
			t.Fatal(err)
		}
		received = append(received, activity)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	delivery := &Delivery{client: server.Client()}
	if err := delivery.DeliverWorkoutDelete("bob", "38472901", []string{server.URL}); err != nil {
		t.Fatalf("DeliverWorkoutDelete() error = %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(received))
	}
	if received[0]["type"] != "Delete" {
		t.Fatalf("type = %v", received[0]["type"])
	}
	object, _ := received[0]["object"].(string)
	if object != workoutObjectURL("bob", "38472901") {
		t.Fatalf("object = %q", object)
	}
}
