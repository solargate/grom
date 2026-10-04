package v1_test

import (
	"net/http"
	"testing"

	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/notifications"
)

func TestPushSubscriptionRegisterUpdateDelete(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, aliceUser := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")
	aliceID, _ := aliceUser["id"].(string)

	body := map[string]any{
		"installation_id": "inst-1",
		"endpoint":        "https://push.example/alice-1",
		"keys": map[string]string{
			"p256dh": "p256-alice",
			"auth":   "auth-alice",
		},
		"platform":    "android",
		"device_name": "Pixel",
	}
	w := ta.doJSON(t, http.MethodPost, "/api/v1/notifications/push", body, aliceToken)
	expectStatus(t, w, http.StatusOK)
	created := decodeObject(t, w)
	if created["installation_id"] != "inst-1" || created["platform"] != "android" || created["device_name"] != "Pixel" {
		t.Fatalf("create response: %#v", created)
	}
	if created["updated_at"] == "" {
		t.Fatalf("missing updated_at: %#v", created)
	}

	list, err := ta.app.PushSubscriptions.ListByUser(aliceID)
	if err != nil || len(list) != 1 || list[0].Endpoint != "https://push.example/alice-1" {
		t.Fatalf("stored: %#v err=%v", list, err)
	}

	body["endpoint"] = "https://push.example/alice-1b"
	w = ta.doJSON(t, http.MethodPost, "/api/v1/notifications/push", body, aliceToken)
	expectStatus(t, w, http.StatusOK)
	list, err = ta.app.PushSubscriptions.ListByUser(aliceID)
	if err != nil || len(list) != 1 || list[0].Endpoint != "https://push.example/alice-1b" {
		t.Fatalf("updated: %#v err=%v", list, err)
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/notifications/push/inst-1", nil, bobToken)
	expectStatus(t, w, http.StatusNotFound)
	list, err = ta.app.PushSubscriptions.ListByUser(aliceID)
	if err != nil || len(list) != 1 {
		t.Fatalf("bob delete must not remove alice sub: %#v err=%v", list, err)
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/notifications/push/inst-1", nil, aliceToken)
	expectStatus(t, w, http.StatusNoContent)
	list, err = ta.app.PushSubscriptions.ListByUser(aliceID)
	if err != nil || len(list) != 0 {
		t.Fatalf("after delete: %#v err=%v", list, err)
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/notifications/push/inst-1", nil, aliceToken)
	expectStatus(t, w, http.StatusNotFound)
}

func TestPushSubscriptionValidationAndAuth(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	jwt, _ := ta.login(t, "alice@example.com", "password12")
	rawPAT, _ := createPAT(t, ta, jwt, "push", []string{pat.ScopeWorkoutsRead}, nil)

	w := ta.doJSON(t, http.MethodPost, "/api/v1/notifications/push", map[string]any{
		"installation_id": "inst-1",
		"endpoint":        "https://push.example/x",
		"keys":            map[string]string{"p256dh": "p", "auth": "a"},
	}, "")
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/notifications/push", map[string]any{
		"installation_id": "inst-1",
		"endpoint":        "https://push.example/x",
		"keys":            map[string]string{"p256dh": "p", "auth": "a"},
	}, rawPAT)
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/notifications/push", map[string]any{
		"installation_id": "  ",
		"endpoint":        "https://push.example/x",
		"keys":            map[string]string{"p256dh": "p", "auth": "a"},
	}, jwt)
	expectStatus(t, w, http.StatusBadRequest)
	if decodeObject(t, w)["error"] != notifications.ErrInvalidSubscription.Error() {
		t.Fatalf("validation error: %s", w.Body.String())
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/notifications/push/inst-1", nil, rawPAT)
	expectStatus(t, w, http.StatusUnauthorized)
}
