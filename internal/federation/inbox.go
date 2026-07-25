package federation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type InboxProcessor struct {
	users          users.Repository
	social         *social.Service
	delivery       *Delivery
	inboxStore     InboxRepository
	followersStore FollowersRepository
	autoAccept     bool
}

func NewInboxProcessor(userStore users.Repository, socialSvc *social.Service, delivery *Delivery, inboxStore InboxRepository, followersStore FollowersRepository) *InboxProcessor {
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
	case "Update":
		return p.handleUpdate(nickname, activity)
	case "Delete":
		return p.handleDelete(nickname, activity)
	case "Undo":
		return p.handleUndo(nickname, activity)
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
		handle := actorToHandle(followerActor)
		if handle == "" {
			handle = followerActor
		}
		_ = p.followersStore.Add(targetNickname, InboundFollower{
			ActorURI: followerActor,
			Inbox:    strings.TrimSuffix(followerActor, "/") + "/inbox",
			Handle:   handle,
		})
		p.cacheInboundFollowerAvatar(targetNickname, handle)
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

func (p *InboxProcessor) cacheInboundFollowerAvatar(targetNickname, handle string) {
	if p.inboxStore == nil || p.delivery == nil || handle == "" {
		return
	}
	parsed := social.ParsedHandle{
		Nickname: ownerNicknameFromDir(OwnerKeyFromHandle(handle)),
		Domain:   domainFromHandle(handle),
		Handle:   handle,
	}
	actor, err := fetchActor(p.delivery.Client(), parsed)
	if err != nil {
		return
	}
	_ = p.inboxStore.EnsureAuthor(
		targetNickname,
		handle,
		parsed.Nickname,
		ExtractActorName(actor),
		ExtractIconURL(actor),
		true,
	)
}

func (p *InboxProcessor) handleUndo(targetNickname string, activity map[string]any) error {
	if p.followersStore == nil {
		return nil
	}
	object, ok := activity["object"].(map[string]any)
	if !ok || object["type"] != "Follow" {
		return nil
	}
	followerActor, _ := object["actor"].(string)
	if followerActor == "" {
		followerActor, _ = activity["actor"].(string)
	}
	if followerActor == "" {
		return nil
	}
	return p.followersStore.Remove(targetNickname, followerActor)
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
	return p.handleWorkoutActivity(viewerNickname, activity, false)
}

func (p *InboxProcessor) handleUpdate(viewerNickname string, activity map[string]any) error {
	return p.handleWorkoutActivity(viewerNickname, activity, true)
}

func (p *InboxProcessor) handleWorkoutActivity(viewerNickname string, activity map[string]any, replace bool) error {
	if p.inboxStore == nil {
		return nil
	}
	object, ok := activity["object"].(map[string]any)
	if !ok {
		return nil
	}
	actorURI, _ := activity["actor"].(string)
	ownerHandle := actorToHandle(actorURI)
	if ownerHandle == "" {
		return nil
	}

	workout, trackData, mediaFiles, err := parseFederatedWorkoutObject(object)
	if err != nil {
		return err
	}
	if workout == nil {
		return nil
	}

	var actorDoc map[string]any
	if p.delivery != nil && actorURI != "" {
		parsed := social.ParsedHandle{
			Nickname: ownerNicknameFromDir(OwnerKeyFromHandle(ownerHandle)),
			Domain:   domainFromHandle(ownerHandle),
			Handle:   ownerHandle,
		}
		if fetched, err := fetchActor(p.delivery.client, parsed); err == nil {
			actorDoc = fetched
		}
	}

	if replace {
		return p.inboxStore.Replace(viewerNickname, ownerHandle, workout, trackData, mediaFiles, actorDoc)
	}
	return p.inboxStore.Save(viewerNickname, ownerHandle, workout, trackData, mediaFiles, actorDoc)
}

func parseFederatedWorkoutObject(object map[string]any) (*workouts.Workout, []byte, []workouts.MediaFileInput, error) {
	workout := workouts.Workout{
		ID:                   workoutIDFromObject(object),
		Name:                 stringValue(object, "name"),
		Description:          stringValue(object, "content"),
		SportType:            stringValue(object, "sportType"),
		Device:               stringValue(object, "device"),
		DurationSeconds:      intValue(object, "durationSeconds"),
		DurationTotalSeconds: intValue(object, "durationTotalSeconds"),
		Distance:             floatValue(object, "distance"),
		Track:                stringValue(object, "track"),
	}
	if start := stringValue(object, "startDate"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			workout.StartDate = t
		}
	}
	if pace := stringValue(object, "tempAvgKmm"); pace != "" {
		workout.TempAvgKmm = &pace
	}
	if v, ok := optionalPositiveFloat(object, "elevationGain"); ok {
		workout.ElevationGain = &v
	}
	if v, ok := optionalPositiveFloat(object, "speedAvgKmh"); ok {
		workout.SpeedAvgKmh = &v
	}
	if v, ok := optionalPositiveFloat(object, "heartRateAvg"); ok {
		workout.HeartRateAvg = &v
	}
	if v, ok := optionalPositiveInt(object, "stepsTotal"); ok {
		workout.StepsTotal = &v
	}
	if v, ok := optionalPositiveFloat(object, "calories"); ok {
		workout.Calories = &v
	}
	if workout.ID == "" || workout.Name == "" {
		return nil, nil, nil, nil
	}

	var trackData []byte
	if workout.Track != "" {
		if encoded := stringValue(object, "trackData"); encoded != "" {
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("decode track data: %w", err)
			}
			if len(decoded) > tracks.MaxTrackSizeBytes {
				return nil, nil, nil, fmt.Errorf("track data too large")
			}
			trackData = decoded
		}
	}

	mediaFiles, err := decodeMediaItems(object)
	if err != nil {
		return nil, nil, nil, err
	}
	return &workout, trackData, mediaFiles, nil
}

func (p *InboxProcessor) handleDelete(viewerNickname string, activity map[string]any) error {
	if p.inboxStore == nil {
		return nil
	}
	actorURI, _ := activity["actor"].(string)
	ownerHandle := actorToHandle(actorURI)
	if ownerHandle == "" {
		return nil
	}

	workoutID := workoutIDFromDeleteObject(activity["object"])
	if workoutID == "" {
		return nil
	}
	return p.inboxStore.Delete(viewerNickname, ownerHandle, workoutID)
}

func workoutIDFromDeleteObject(raw any) string {
	switch object := raw.(type) {
	case string:
		return workoutIDFromObject(map[string]any{"id": object})
	case map[string]any:
		if id := workoutIDFromObject(object); id != "" {
			return id
		}
		if nested, ok := object["object"].(map[string]any); ok {
			return workoutIDFromObject(nested)
		}
	}
	return ""
}

func domainFromHandle(handle string) string {
	if idx := strings.LastIndex(handle, "@"); idx >= 0 && idx < len(handle)-1 {
		return handle[idx+1:]
	}
	return ""
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

func decodeMediaItems(object map[string]any) ([]workouts.MediaFileInput, error) {
	rawItems, ok := object["mediaItems"].([]any)
	if !ok || len(rawItems) == 0 {
		return nil, nil
	}
	if len(rawItems) > workouts.MaxPhotosPerWorkout {
		return nil, fmt.Errorf("too many media items")
	}
	files := make([]workouts.MediaFileInput, 0, len(rawItems))
	for _, raw := range rawItems {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		filename := stringValue(item, "filename")
		encoded := stringValue(item, "data")
		if filename == "" || encoded == "" {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("decode media data: %w", err)
		}
		if len(data) > workouts.MaxPhotoBytes {
			return nil, workouts.ErrPhotoTooLarge
		}
		files = append(files, workouts.MediaFileInput{
			Filename: filename,
			Data:     data,
		})
	}
	return files, nil
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

func optionalPositiveFloat(m map[string]any, key string) (float64, bool) {
	raw, ok := m[key]
	if !ok || raw == nil {
		return 0, false
	}
	var v float64
	switch n := raw.(type) {
	case float64:
		v = n
	case int:
		v = float64(n)
	default:
		return 0, false
	}
	if v <= 0 {
		return 0, false
	}
	return v, true
}

func optionalPositiveInt(m map[string]any, key string) (int, bool) {
	raw, ok := m[key]
	if !ok || raw == nil {
		return 0, false
	}
	var v int
	switch n := raw.(type) {
	case float64:
		v = int(n)
	case int:
		v = n
	default:
		return 0, false
	}
	if v <= 0 {
		return 0, false
	}
	return v, true
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
