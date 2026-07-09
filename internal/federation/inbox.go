package federation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solargate/travka/internal/config"
	"github.com/solargate/travka/internal/social"
	"github.com/solargate/travka/internal/users"
	"github.com/solargate/travka/internal/workouts"
)

type InboxProcessor struct {
	users          *users.Store
	social         *social.Service
	delivery       *Delivery
	inboxStore     *WorkoutInboxStore
	followersStore *FollowersStore
	autoAccept     bool
}

func NewInboxProcessor(userStore *users.Store, socialSvc *social.Service, delivery *Delivery, inboxStore *WorkoutInboxStore, followersStore *FollowersStore) *InboxProcessor {
	return &InboxProcessor{
		users:          userStore,
		social:         socialSvc,
		delivery:       delivery,
		inboxStore:     inboxStore,
		followersStore: followersStore,
		autoAccept:     config.Cfg.Federation.AutoAcceptFollows,
	}
}

func (p *InboxProcessor) Handle(nickname string, body io.Reader) error {
	var activity map[string]any
	if err := json.NewDecoder(body).Decode(&activity); err != nil {
		return err
	}

	actType, _ := activity["type"].(string)
	switch actType {
	case "Follow":
		return p.handleFollow(nickname, activity)
	case "Accept":
		return p.handleAccept(nickname, activity)
	case "Create":
		return p.handleCreate(nickname, activity)
	case "Undo":
		return nil
	default:
		return nil
	}
}

func (p *InboxProcessor) handleFollow(targetNickname string, activity map[string]any) error {
	followerActor, _ := activity["actor"].(string)
	if followerActor == "" {
		return fmt.Errorf("missing actor")
	}
	followID, _ := activity["id"].(string)

	if p.followersStore != nil {
		_ = p.followersStore.Add(targetNickname, InboundFollower{
			ActorURI: followerActor,
			Inbox:    strings.TrimSuffix(followerActor, "/") + "/inbox",
			Handle:   followerActor,
		})
	}

	if p.autoAccept && p.delivery != nil {
		targetActor := actorURL(targetNickname)
		accept := map[string]any{
			"@context": "https://www.w3.org/ns/activitystreams",
			"id":       fmt.Sprintf("%s/accepts/%s", targetActor, uuid.NewString()),
			"type":     "Accept",
			"actor":    targetActor,
			"object": map[string]any{
				"id":     followID,
				"type":   "Follow",
				"actor":  followerActor,
				"object": targetActor,
			},
		}
		inbox := strings.TrimSuffix(followerActor, "/") + "/inbox"
		return p.delivery.postActivity(inbox, accept)
	}
	return nil
}

func (p *InboxProcessor) handleAccept(viewerNickname string, activity map[string]any) error {
	object, ok := activity["object"].(map[string]any)
	if !ok {
		return nil
	}
	followID, _ := object["id"].(string)
	if followID == "" || p.social == nil {
		return nil
	}
	return p.social.ActivateFollowByActivityID(followID)
}

func (p *InboxProcessor) handleCreate(viewerNickname string, activity map[string]any) error {
	if p.inboxStore == nil {
		return nil
	}
	object, ok := activity["object"].(map[string]any)
	if !ok {
		return nil
	}
	actor, _ := activity["actor"].(string)
	ownerHandle := actorToHandle(actor)
	if ownerHandle == "" {
		return nil
	}

	workout := workouts.Workout{
		ID:              workoutIDFromObject(object),
		Name:            stringValue(object, "name"),
		Description:     stringValue(object, "content"),
		SportType:       stringValue(object, "sportType"),
		DurationSeconds: intValue(object, "durationSeconds"),
		Distance:        floatValue(object, "distance"),
		Track:           stringValue(object, "track"),
	}
	if start := stringValue(object, "startDate"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			workout.StartDate = t
		}
	}
	if workout.ID == "" || workout.Name == "" {
		return nil
	}
	return p.inboxStore.Save(viewerNickname, ownerHandle, &workout)
}

func actorToHandle(actor string) string {
	actor = strings.TrimSuffix(actor, "/")
	if !strings.HasPrefix(actor, "https://") {
		return ""
	}
	rest := strings.TrimPrefix(actor, "https://")
	idx := strings.Index(rest, "/users/")
	if idx < 0 {
		return ""
	}
	domain := rest[:idx]
	nick := strings.TrimPrefix(rest[idx+len("/users/"):], "/")
	nick = strings.TrimSuffix(nick, "/")
	if nick == "" || domain == "" {
		return ""
	}
	return nick + "@" + domain
}

func workoutIDFromObject(object map[string]any) string {
	raw := stringValue(object, "id")
	if raw == "" {
		return ""
	}
	if idx := strings.LastIndex(raw, "/"); idx >= 0 && idx+1 < len(raw) {
		return raw[idx+1:]
	}
	return raw
}

func stringValue(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func intValue(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func floatValue(m map[string]any, key string) float64 {
	v, _ := m[key].(float64)
	return v
}

func (d *Delivery) postActivity(inbox string, activity map[string]any) error {
	body, err := json.Marshal(activity)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, inbox, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/activity+json")
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("activity post failed: %s", resp.Status)
	}
	return nil
}
