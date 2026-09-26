package federation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/workouts"
)

// ParseWorkoutObject parses a Grom ActivityPub Workout object.
// Returns nil workout (no error) when the object is not a Grom workout.
func ParseWorkoutObject(object map[string]any) (*workouts.Workout, []byte, []workouts.MediaFileInput, error) {
	return parseFederatedWorkoutObject(object)
}

// FetchJSONSigned GETs url with Accept activity+json and optional instance signature.
func FetchJSONSigned(client *http.Client, blobs blob.Store, url string) (map[string]any, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/activity+json")
	if err := signOutboundGET(blobs, req); err != nil {
		// Best-effort unsigned when keys unavailable.
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote fetch %s: status %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// FetchWorkoutObject loads a remote Workout object by ActivityPub id.
func FetchWorkoutObject(client *http.Client, blobs blob.Store, objectURL string) (*workouts.Workout, []byte, []workouts.MediaFileInput, string, error) {
	doc, err := FetchJSONSigned(client, blobs, objectURL)
	if err != nil {
		return nil, nil, nil, "", err
	}
	// Unwrap Create activity if needed.
	if typ, _ := doc["type"].(string); typ == "Create" {
		if obj, ok := doc["object"].(map[string]any); ok {
			doc = obj
		}
	}
	workout, track, media, err := parseFederatedWorkoutObject(doc)
	if err != nil {
		return nil, nil, nil, "", err
	}
	if workout == nil {
		return nil, nil, nil, "", fmt.Errorf("not a grom workout object")
	}
	objectID := stringValue(doc, "id")
	if objectID == "" {
		objectID = objectURL
	}
	return workout, track, media, objectID, nil
}

// OutboxWorkout is a workout discovered in a remote outbox page.
type OutboxWorkout struct {
	Workout  *workouts.Workout
	ObjectID string
	Track    []byte
	Media    []workouts.MediaFileInput
}

// FetchOutboxWorkouts loads Grom workouts from a remote actor outbox (first pages).
func FetchOutboxWorkouts(client *http.Client, blobs blob.Store, parsed social.ParsedHandle, limit int) ([]OutboxWorkout, error) {
	actor, err := fetchActor(client, blobs, parsed)
	if err != nil {
		return nil, err
	}
	outboxURL, _ := actor["outbox"].(string)
	if outboxURL == "" {
		return nil, nil
	}
	collection, err := FetchJSONSigned(client, blobs, outboxURL)
	if err != nil {
		return nil, err
	}
	pageURL := outboxURL
	if first, ok := collection["first"].(string); ok && first != "" {
		pageURL = first
	} else if firstObj, ok := collection["first"].(map[string]any); ok {
		if id, _ := firstObj["id"].(string); id != "" {
			pageURL = id
		}
	}

	var items []OutboxWorkout
	for pageURL != "" && (limit <= 0 || len(items) < limit) {
		page, err := FetchJSONSigned(client, blobs, pageURL)
		if err != nil {
			return items, err
		}
		ordered, _ := page["orderedItems"].([]any)
		if ordered == nil {
			ordered, _ = page["items"].([]any)
		}
		for _, raw := range ordered {
			obj, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if typ, _ := obj["type"].(string); typ == "Create" {
				if inner, ok := obj["object"].(map[string]any); ok {
					obj = inner
				}
			}
			if typ, _ := obj["type"].(string); typ != "Workout" {
				continue
			}
			workout, track, media, err := parseFederatedWorkoutObject(obj)
			if err != nil || workout == nil {
				continue
			}
			objectID := stringValue(obj, "id")
			items = append(items, OutboxWorkout{
				Workout:  workout,
				ObjectID: objectID,
				Track:    track,
				Media:    media,
			})
			if limit > 0 && len(items) >= limit {
				break
			}
		}
		next, _ := page["next"].(string)
		pageURL = next
		if pageURL == outboxURL {
			break
		}
	}
	return items, nil
}

// FetchActorCollectionItems loads actor URLs/handles from an AP collection (followers/following).
func FetchActorCollectionItems(client *http.Client, blobs blob.Store, parsed social.ParsedHandle, field string, limit int) ([]social.UserSearchResult, error) {
	actor, err := fetchActor(client, blobs, parsed)
	if err != nil {
		return nil, err
	}
	collURL, _ := actor[field].(string)
	if collURL == "" {
		return nil, nil
	}
	collection, err := FetchJSONSigned(client, blobs, collURL)
	if err != nil {
		return nil, err
	}
	pageURL := collURL
	if first, ok := collection["first"].(string); ok && first != "" {
		pageURL = first
	}
	var out []social.UserSearchResult
	for pageURL != "" && (limit <= 0 || len(out) < limit) {
		page, err := FetchJSONSigned(client, blobs, pageURL)
		if err != nil {
			return out, err
		}
		ordered, _ := page["orderedItems"].([]any)
		if ordered == nil {
			ordered, _ = page["items"].([]any)
		}
		for _, raw := range ordered {
			actorURI := ""
			switch v := raw.(type) {
			case string:
				actorURI = v
			case map[string]any:
				if id, _ := v["id"].(string); id != "" {
					actorURI = id
				} else if href, _ := v["href"].(string); href != "" {
					actorURI = href
				}
			}
			if actorURI == "" {
				continue
			}
			handle := actorURIToHandle(actorURI)
			if handle == "" {
				continue
			}
			nick := handle
			if at := strings.LastIndex(handle, "@"); at > 0 {
				nick = handle[:at]
			}
			name := ""
			if m, ok := raw.(map[string]any); ok {
				name = ExtractActorName(m)
			}
			out = append(out, social.UserSearchResult{
				Nickname: nick,
				Name:     name,
				Handle:   handle,
				IsLocal:  false,
			})
			if limit > 0 && len(out) >= limit {
				break
			}
		}
		next, _ := page["next"].(string)
		pageURL = next
	}
	return out, nil
}

func actorURIToHandle(actorURI string) string {
	actorURI = strings.TrimSpace(actorURI)
	actorURI = strings.TrimSuffix(actorURI, "/")
	// https://host/users/nick
	const marker = "/users/"
	idx := strings.Index(actorURI, marker)
	if idx < 0 {
		return ""
	}
	rest := actorURI[idx+len(marker):]
	if slash := strings.Index(rest, "/"); slash >= 0 {
		rest = rest[:slash]
	}
	if rest == "" {
		return ""
	}
	host := ""
	if u := strings.TrimPrefix(actorURI, "https://"); u != actorURI {
		if slash := strings.Index(u, "/"); slash > 0 {
			host = u[:slash]
		}
	} else if u := strings.TrimPrefix(actorURI, "http://"); u != actorURI {
		if slash := strings.Index(u, "/"); slash > 0 {
			host = u[:slash]
		}
	}
	if host == "" {
		return rest
	}
	return rest + "@" + host
}
