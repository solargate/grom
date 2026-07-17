package federation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
)

func TestResolveRemoteViaWebFinger(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Enabled = true
	config.Cfg.Federation.Domain = "grom.test"

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/webfinger", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"links": []map[string]any{{
				"rel":  "self",
				"href": "https://" + r.Host + "/users/bob",
			}},
		})
	})
	mux.HandleFunc("/users/bob", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/activity+json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "Person",
			"name": "Bob Remote",
			"icon": map[string]any{"url": "https://cdn.example/bob.png"},
		})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	host := server.Listener.Addr().String()
	delivery := &Delivery{client: server.Client()}
	parsed := social.ParsedHandle{
		Nickname: "bob",
		Domain:   host,
		Handle:   "bob@" + host,
	}
	res, err := delivery.ResolveRemote(parsed)
	if err != nil {
		t.Fatalf("ResolveRemote: %v", err)
	}
	if res.Name != "Bob Remote" {
		t.Fatalf("name = %q", res.Name)
	}
	if res.AvatarURL != "https://cdn.example/bob.png" {
		t.Fatalf("avatar = %q", res.AvatarURL)
	}
}
