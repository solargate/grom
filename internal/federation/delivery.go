package federation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/server"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type Delivery struct {
	client    *http.Client
	userStore users.Repository
	social    *social.Service
}

func NewDelivery(userStore users.Repository, socialSvc *social.Service) (*Delivery, error) {
	client, err := server.FederationHTTPClient()
	if err != nil {
		return nil, err
	}
	return &Delivery{
		client:    client,
		userStore: userStore,
		social:    socialSvc,
	}, nil
}

func (d *Delivery) Client() *http.Client {
	return d.client
}

func (d *Delivery) DeliverFollow(follow *social.Follow) error {
	follower, err := d.userStore.FindByID(follow.FollowerID)
	if err != nil {
		return err
	}

	activityID := follow.FollowActivityID
	if activityID == "" {
		activityID = fmt.Sprintf("%s/follows/%s", actorURL(follower.Nickname), uuid.NewString())
		follow.FollowActivityID = activityID
	}

	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       activityID,
		"type":     "Follow",
		"actor":    actorURL(follower.Nickname),
		"object":   follow.TargetActorURI,
	}
	body, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	inbox := strings.TrimSuffix(follow.TargetActorURI, "/") + "/inbox"
	req, err := http.NewRequest(http.MethodPost, inbox, bytes.NewReader(body))
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
		return fmt.Errorf("follow delivery failed: %s", resp.Status)
	}
	return nil
}

func (d *Delivery) DeliverUndo(follow *social.Follow) error {
	follower, err := d.userStore.FindByID(follow.FollowerID)
	if err != nil {
		return err
	}
	if follow.FollowActivityID == "" {
		return nil
	}

	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       fmt.Sprintf("%s/undos/%s", actorURL(follower.Nickname), uuid.NewString()),
		"type":     "Undo",
		"actor":    actorURL(follower.Nickname),
		"object": map[string]any{
			"id":     follow.FollowActivityID,
			"type":   "Follow",
			"actor":  actorURL(follower.Nickname),
			"object": follow.TargetActorURI,
		},
	}
	body, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	inbox := strings.TrimSuffix(follow.TargetActorURI, "/") + "/inbox"
	req, err := http.NewRequest(http.MethodPost, inbox, bytes.NewReader(body))
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
		return fmt.Errorf("undo delivery failed: %s", resp.Status)
	}
	return nil
}

func (d *Delivery) ResolveRemote(parsed social.ParsedHandle) (*social.UserSearchResult, error) {
	if !config.Cfg.Federation.Enabled {
		return nil, social.ErrRemoteNotReady
	}
	actor, err := fetchActor(d.client, parsed)
	if err != nil {
		return nil, err
	}
	name := ExtractActorName(actor)
	avatarURL := ExtractIconURL(actor)
	return &social.UserSearchResult{
		Nickname:  parsed.Nickname,
		Name:      name,
		Handle:    parsed.Handle,
		IsLocal:   false,
		HasAvatar: avatarURL != "",
		AvatarURL: avatarURL,
	}, nil
}

func fetchActor(client *http.Client, parsed social.ParsedHandle) (map[string]any, error) {
	webfingerURL := fmt.Sprintf("https://%s/.well-known/webfinger?resource=acct:%s",
		parsed.Domain, parsed.Handle)
	req, err := http.NewRequest(http.MethodGet, webfingerURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, social.ErrUserNotFound
	}
	var jrd map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&jrd); err != nil {
		return nil, err
	}
	links, _ := jrd["links"].([]any)
	var actorURL string
	for _, link := range links {
		m, _ := link.(map[string]any)
		if m["rel"] == "self" {
			actorURL, _ = m["href"].(string)
			break
		}
	}
	if actorURL == "" {
		return nil, social.ErrUserNotFound
	}

	actorReq, err := http.NewRequest(http.MethodGet, actorURL, nil)
	if err != nil {
		return nil, err
	}
	actorReq.Header.Set("Accept", "application/activity+json")
	actorResp, err := client.Do(actorReq)
	if err != nil {
		return nil, err
	}
	defer actorResp.Body.Close()
	if actorResp.StatusCode != http.StatusOK {
		return nil, social.ErrUserNotFound
	}
	var actor map[string]any
	if err := json.NewDecoder(actorResp.Body).Decode(&actor); err != nil {
		return nil, err
	}
	return actor, nil
}

func (d *Delivery) DeliverWorkout(authorNickname string, workout *workouts.Workout, followerInboxes []string, trackData []byte, mediaFiles []workouts.MediaFileInput) error {
	if len(followerInboxes) == 0 {
		return nil
	}
	author := actorURL(authorNickname)
	object := map[string]any{
		"id":              workoutObjectURL(authorNickname, workout.ID),
		"type":            "Workout",
		"name":            workout.Name,
		"content":         workout.Description,
		"sportType":       workout.SportType,
		"startDate":       workout.StartDate.UTC().Format(time.RFC3339),
		"device":          workout.Device,
		"durationSeconds": workout.DurationSeconds,
		"distance":        workout.Distance,
		"track":           workout.Track,
	}
	if workout.Track != "" && len(trackData) > 0 {
		object["trackData"] = base64.StdEncoding.EncodeToString(trackData)
	}
	if len(mediaFiles) > 0 {
		items := make([]map[string]any, 0, len(mediaFiles))
		for _, file := range mediaFiles {
			items = append(items, map[string]any{
				"filename":  file.Filename,
				"mediaType": workouts.MediaContentType(file.Filename),
				"data":      base64.StdEncoding.EncodeToString(file.Data),
			})
		}
		object["mediaItems"] = items
	}
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       fmt.Sprintf("%s/activities/%s", author, uuid.NewString()),
		"type":     "Create",
		"actor":    author,
		"object":   object,
		"to":       []string{"https://www.w3.org/ns/activitystreams#Public"},
	}
	for _, inbox := range followerInboxes {
		if err := d.postActivity(inbox, activity); err != nil {
			return err
		}
	}
	return nil
}

func (d *Delivery) DeliverWorkoutDelete(authorNickname, workoutID string, followerInboxes []string) error {
	if len(followerInboxes) == 0 {
		return nil
	}
	author := actorURL(authorNickname)
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       fmt.Sprintf("%s/activities/%s", author, uuid.NewString()),
		"type":     "Delete",
		"actor":    author,
		"object":   workoutObjectURL(authorNickname, workoutID),
		"to":       []string{"https://www.w3.org/ns/activitystreams#Public"},
	}
	for _, inbox := range followerInboxes {
		if err := d.postActivity(inbox, activity); err != nil {
			return err
		}
	}
	return nil
}
