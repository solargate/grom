package v1_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/federation"
)

func TestDeleteAccountRequiresPasswordAndPurges(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Morning run",
		"sport_type":       "Run",
		"start_date":       time.Now().UTC().Format(time.RFC3339),
		"duration_seconds": 600,
	}, token)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "wrong-password",
	}, token)
	expectStatus(t, w, http.StatusForbidden)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, token)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, token)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, token)
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusUnauthorized)

	userDir := data.UserDir(ta.dir, "alice")
	if _, err := os.Stat(userDir); !os.IsNotExist(err) {
		t.Fatalf("expected user dir removed, stat err=%v path=%s", err, userDir)
	}

	registered := ta.register(t, "alice", "alice@example.com", "password12")
	if registered["nickname"] != "alice" {
		t.Fatalf("re-register failed: %#v", registered)
	}
}

func TestDeleteAccountRejectsEmptyPassword(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "",
	}, token)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "   ",
	}, token)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", nil, token)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, token)
	expectStatus(t, w, http.StatusOK)
}

func TestDeleteAccountRequiresAuth(t *testing.T) {
	ta := setupTestApp(t)
	w := ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestDeleteAccountRejectsPAT(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
		"name":   "cli",
		"scopes": []string{"workouts:read"},
	}, token)
	expectStatus(t, w, http.StatusCreated)
	patToken, _ := decodeObject(t, w)["token"].(string)
	if patToken == "" {
		t.Fatal("expected pat token")
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, patToken)
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestDeleteAccountRemovesLikesOnOthersWorkouts(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Alice run",
		"sport_type":       "Run",
		"start_date":       time.Now().UTC().Format(time.RFC3339),
		"duration_seconds": 600,
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "alice",
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+workoutID+"/likes?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, bobToken)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+workoutID+"/likes?owner=alice", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	likes := decodeObject(t, w)
	if likes["count"] != float64(0) {
		t.Fatalf("expected likes cleared, got %#v", likes)
	}
}

func TestDeleteAccountRemovesCommentsOnOthersWorkouts(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Alice run",
		"sport_type":       "Run",
		"start_date":       time.Now().UTC().Format(time.RFC3339),
		"duration_seconds": 600,
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "alice",
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	path := "/api/v1/workouts/" + workoutID + "/comments?owner=alice"
	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": "Nice run"}, bobToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(1) {
		t.Fatalf("comment create: %#v", decodeObject(t, w))
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, bobToken)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	comments := decodeObject(t, w)
	if comments["count"] != float64(0) {
		t.Fatalf("expected comments cleared, got %#v", comments)
	}
}

func TestDeleteAccountRemovesFollowsBothWays(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "bob",
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "alice",
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, bobToken)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/social/following", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	following := decodeList(t, w)
	if len(following) != 0 {
		t.Fatalf("expected alice following cleared of bob, got %#v", following)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/social/followers", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	followers := decodeList(t, w)
	if len(followers) != 0 {
		t.Fatalf("expected alice followers cleared of bob, got %#v", followers)
	}
}

func TestDeleteAccountPurgesOnFileAndBBolt(t *testing.T) {
	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			ta := setupTestAppWithDriver(t, driver)
			ta.register(t, "alice", "alice@example.com", "password12")
			token, user := ta.login(t, "alice@example.com", "password12")
			userID, _ := user["id"].(string)

			w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
				"name":             "Morning run",
				"sport_type":       "Run",
				"start_date":       time.Now().UTC().Format(time.RFC3339),
				"duration_seconds": 600,
			}, token)
			expectStatus(t, w, http.StatusCreated)
			workoutID, _ := decodeObject(t, w)["id"].(string)

			w = ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
				"type": "shoes",
				"name": "Trail",
			}, token)
			expectStatus(t, w, http.StatusCreated)
			eqID, _ := decodeObject(t, w)["id"].(string)

			w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
				"name":   "cli",
				"scopes": []string{"workouts:read"},
			}, token)
			expectStatus(t, w, http.StatusCreated)

			w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
				"password": "password12",
			}, token)
			expectStatus(t, w, http.StatusNoContent)

			if _, err := ta.app.Workouts.Get("alice", workoutID); err == nil {
				t.Fatal("expected workout removed")
			}
			if _, err := ta.app.Equipment.FindByID("alice", eqID); err == nil {
				t.Fatal("expected equipment removed")
			}
			pats, err := ta.app.Backend.PAT().ListByUser(userID)
			if err != nil {
				t.Fatal(err)
			}
			if len(pats) != 0 {
				t.Fatalf("expected PATs purged, got %#v", pats)
			}
			if _, err := os.Stat(data.UserDir(ta.dir, "alice")); !os.IsNotExist(err) {
				t.Fatalf("expected user dir removed, err=%v", err)
			}

			registered := ta.register(t, "alice", "alice@example.com", "password12")
			if registered["nickname"] != "alice" {
				t.Fatalf("re-register failed: %#v", registered)
			}
			newToken, _ := ta.login(t, "alice@example.com", "password12")
			w = ta.doJSON(t, http.MethodGet, "/api/v1/equipment", nil, newToken)
			expectStatus(t, w, http.StatusOK)
			if len(decodeList(t, w)) != 0 {
				t.Fatal("expected empty equipment after re-register")
			}
		})
	}
}

func TestDeleteAccountDeliversActorDeleteBestEffort(t *testing.T) {
	ta := setupFederationTestApp(t)
	transport := &captureDeliveryRT{}
	ta.app.SetFederationHTTPClient(&http.Client{Transport: transport})

	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	actorURI := "https://remote.example/users/bob"
	seedRemoteFollow(t, ta, "alice", "bob@remote.example", actorURI)
	if err := ta.app.Federation.Followers().Add("alice", federation.InboundFollower{
		ActorURI: "https://remote.example/users/carol",
		Inbox:    "https://remote.example/users/carol/inbox",
		Handle:   "carol@remote.example",
	}); err != nil {
		t.Fatal(err)
	}

	w := ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, token)
	expectStatus(t, w, http.StatusNoContent)

	acts := transport.activities()
	actorDeletes := 0
	for _, act := range acts {
		if act["type"] != "Delete" {
			continue
		}
		obj, _ := act["object"].(string)
		if obj == "https://localhost/users/alice" {
			actorDeletes++
		}
	}
	if actorDeletes < 1 {
		t.Fatalf("expected actor Delete delivery, got %#v", acts)
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestDeleteAccountSucceedsWhenRemoteInboxFails(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.app.SetFederationHTTPClient(&http.Client{Transport: failingDeliveryRT{}})

	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")
	seedRemoteFollow(t, ta, "alice", "bob@remote.example", "https://remote.example/users/bob")

	w := ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, token)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestFederationInboxAppliesRemoteActorDelete(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	actorURL := "https://remote.example/users/bob"
	handle := "bob@remote.example"
	seedRemoteFollow(t, ta, "alice", handle, actorURL)
	workoutID := "44444444"
	postFederatedWorkoutCreate(t, ta, "alice", actorURL, workoutID, map[string]any{
		"likesCount": 1,
		"likedUsers": []any{
			map[string]any{
				"handle": "carol@other.test", "nickname": "carol", "name": "Carol", "is_local": false,
			},
		},
		"commentsCount": 1,
		"comments": []any{
			map[string]any{
				"id": "c1", "datetime": "2026-08-06T12:00:00Z", "text": "Hi",
				"noteId": "https://other.test/users/carol/notes/c1",
				"user": map[string]any{
					"handle": "carol@other.test", "nickname": "carol", "name": "Carol", "is_local": false,
				},
			},
		},
	})

	likes, err := ta.app.Likes.GetFederated("alice", handle, workoutID)
	if err != nil || likes.Likes != 1 {
		t.Fatalf("seed likes: %#v err=%v", likes, err)
	}
	comments, err := ta.app.Comments.GetFederated("alice", handle, workoutID)
	if err != nil || comments.CommentsNum != 1 {
		t.Fatalf("seed comments: %#v err=%v", comments, err)
	}

	deleteObj := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Delete",
		"actor":    actorURL,
		"object":   actorURL,
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	keyID := actorURL + "#main-key"
	ta.installRemoteKey(t, priv, keyID, actorURL)
	w := ta.postSignedActivity(t, "/users/alice/inbox", deleteObj, priv, keyID)
	expectStatus(t, w, http.StatusAccepted)

	items, err := ta.app.Federation.Inbox().List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected inbox purged, got %#v", items)
	}
	likes, err = ta.app.Likes.GetFederated("alice", handle, workoutID)
	if err != nil || likes.Likes != 0 {
		t.Fatalf("expected federated likes cleared: %#v err=%v", likes, err)
	}
	comments, err = ta.app.Comments.GetFederated("alice", handle, workoutID)
	if err != nil || comments.CommentsNum != 0 {
		t.Fatalf("expected federated comments cleared: %#v err=%v", comments, err)
	}

	following, err := ta.app.Social.ListFollowing(mustUserID(t, ta, "alice"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range following {
		if f.TargetHandle == handle {
			t.Fatalf("expected follow to remote actor removed: %#v", following)
		}
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=feed&limit=20", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	feedItems, _, _ := decodeWorkoutPage(t, w)
	for _, item := range feedItems {
		if item["id"] == workoutID {
			t.Fatalf("federated workout still in feed: %#v", item)
		}
	}
}

func mustUserID(t *testing.T, ta *testApp, nickname string) string {
	t.Helper()
	u, err := ta.app.Users.FindByNickname(nickname)
	if err != nil {
		t.Fatal(err)
	}
	return u.ID
}

type failingDeliveryRT struct{}

func (failingDeliveryRT) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodPost {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       http.NoBody,
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	return (&captureDeliveryRT{}).RoundTrip(req)
}
