package federation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation/httpsig"
	"github.com/solargate/grom/internal/server"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type Delivery struct {
	client    *http.Client
	userStore users.Repository
	social    *social.Service
	blobs     blob.Store
}

func NewDelivery(userStore users.Repository, socialSvc *social.Service, blobs blob.Store) (*Delivery, error) {
	client, err := server.FederationHTTPClient()
	if err != nil {
		return nil, err
	}
	if err := ensureInstanceActorKey(blobs); err != nil {
		return nil, fmt.Errorf("instance actor key: %w", err)
	}
	return &Delivery{
		client:    client,
		userStore: userStore,
		social:    socialSvc,
		blobs:     blobs,
	}, nil
}

func (d *Delivery) Client() *http.Client {
	return d.client
}

// SetClient replaces the HTTP client used for outbound ActivityPub delivery.
// Intended for tests that capture or stub remote inbox POSTs.
func (d *Delivery) SetClient(client *http.Client) {
	if d == nil || client == nil {
		return
	}
	d.client = client
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
	inbox := strings.TrimSuffix(follow.TargetActorURI, "/") + "/inbox"
	if d.blobs != nil {
		if resolved := d.resolveInboxURL(follow.TargetActorURI); resolved != "" {
			inbox = resolved
		}
	}
	if err := d.postActivity(inbox, activity); err != nil {
		slog.Error("federation follow delivery failed", "inbox", inbox, "err", err)
		return err
	}
	slog.Info("federation follow delivered", "inbox", inbox)
	return nil
}

func (d *Delivery) resolveInboxURL(actorURI string) string {
	parsed, err := parseActorURI(actorURI)
	if err != nil {
		return ""
	}
	actor, err := fetchActor(d.client, d.blobs, parsed)
	if err != nil {
		return ""
	}
	return PreferDeliveryInbox(ExtractActorEndpoints(actor))
}

func parseActorURI(actorURI string) (social.ParsedHandle, error) {
	actorURI = strings.TrimSpace(actorURI)
	const marker = "/users/"
	idx := strings.Index(actorURI, marker)
	if idx < 0 {
		return social.ParsedHandle{}, fmt.Errorf("not a user actor URL")
	}
	hostPart := strings.TrimPrefix(actorURI[:idx], "https://")
	hostPart = strings.TrimPrefix(hostPart, "http://")
	nick := actorURI[idx+len(marker):]
	nick = strings.SplitN(nick, "/", 2)[0]
	nick = strings.SplitN(nick, "#", 2)[0]
	if hostPart == "" || nick == "" {
		return social.ParsedHandle{}, fmt.Errorf("invalid actor URL")
	}
	return social.ParsedHandle{
		Nickname: nick,
		Domain:   hostPart,
		Handle:   nick + "@" + hostPart,
	}, nil
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
	inbox := strings.TrimSuffix(follow.TargetActorURI, "/") + "/inbox"
	if d.blobs != nil {
		if resolved := d.resolveInboxURL(follow.TargetActorURI); resolved != "" {
			inbox = resolved
		}
	}
	if err := d.postActivity(inbox, activity); err != nil {
		slog.Error("federation undo delivery failed", "inbox", inbox, "err", err)
		return err
	}
	slog.Info("federation undo delivered", "inbox", inbox)
	return nil
}

func remoteActorURLFromHandle(handle string) string {
	idx := strings.LastIndex(handle, "@")
	if idx <= 0 || idx >= len(handle)-1 {
		return ""
	}
	nickname := handle[:idx]
	domain := handle[idx+1:]
	return fmt.Sprintf("https://%s/users/%s", domain, nickname)
}

func (d *Delivery) DeliverWorkoutLike(actorNickname, targetHandle, objectID string) (string, error) {
	targetActor := remoteActorURLFromHandle(targetHandle)
	if targetActor == "" {
		return "", fmt.Errorf("invalid target handle")
	}
	activityID := fmt.Sprintf("%s/activities/%s", actorURL(actorNickname), uuid.NewString())
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       activityID,
		"type":     "Like",
		"actor":    actorURL(actorNickname),
		"object":   objectID,
		"to":       []string{"https://www.w3.org/ns/activitystreams#Public", targetActor},
	}
	inbox := strings.TrimSuffix(targetActor, "/") + "/inbox"
	if d.blobs != nil {
		if resolved := d.resolveInboxURL(targetActor); resolved != "" {
			inbox = resolved
		}
	}
	if err := d.postActivity(inbox, activity); err != nil {
		return "", err
	}
	return activityID, nil
}

func (d *Delivery) DeliverWorkoutUndoLike(actorNickname, targetHandle, objectID, likeActivityID string) error {
	targetActor := remoteActorURLFromHandle(targetHandle)
	if targetActor == "" {
		return fmt.Errorf("invalid target handle")
	}
	if likeActivityID == "" {
		likeActivityID = fmt.Sprintf("%s/activities/%s", actorURL(actorNickname), uuid.NewString())
	}
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       fmt.Sprintf("%s/undos/%s", actorURL(actorNickname), uuid.NewString()),
		"type":     "Undo",
		"actor":    actorURL(actorNickname),
		"object": map[string]any{
			"id":     likeActivityID,
			"type":   "Like",
			"actor":  actorURL(actorNickname),
			"object": objectID,
		},
		"to": []string{"https://www.w3.org/ns/activitystreams#Public", targetActor},
	}
	inbox := strings.TrimSuffix(targetActor, "/") + "/inbox"
	if d.blobs != nil {
		if resolved := d.resolveInboxURL(targetActor); resolved != "" {
			inbox = resolved
		}
	}
	return d.postActivity(inbox, activity)
}

func (d *Delivery) DeliverWorkoutComment(actorNickname, targetHandle, workoutObjectID, noteID, text string, published time.Time) (string, error) {
	targetActor := remoteActorURLFromHandle(targetHandle)
	if targetActor == "" {
		return "", fmt.Errorf("invalid target handle")
	}
	if noteID == "" {
		return "", fmt.Errorf("note id required")
	}
	if published.IsZero() {
		published = time.Now().UTC()
	}
	actor := actorURL(actorNickname)
	activityID := fmt.Sprintf("%s/activities/%s", actor, uuid.NewString())
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       activityID,
		"type":     "Create",
		"actor":    actor,
		"object": map[string]any{
			"id":           noteID,
			"type":         "Note",
			"attributedTo": actor,
			"content":      text,
			"published":    published.UTC().Format(time.RFC3339),
			"inReplyTo":    workoutObjectID,
			"to":           []string{"https://www.w3.org/ns/activitystreams#Public"},
		},
		"to": []string{"https://www.w3.org/ns/activitystreams#Public", targetActor},
	}
	inbox := strings.TrimSuffix(targetActor, "/") + "/inbox"
	if d.blobs != nil {
		if resolved := d.resolveInboxURL(targetActor); resolved != "" {
			inbox = resolved
		}
	}
	if err := d.postActivity(inbox, activity); err != nil {
		return "", err
	}
	return activityID, nil
}

func (d *Delivery) DeliverWorkoutCommentDelete(actorNickname, targetHandle, noteID string) error {
	return d.DeliverWorkoutCommentDeleteWithReply(actorNickname, targetHandle, noteID, "")
}

func (d *Delivery) DeliverWorkoutCommentDeleteWithReply(actorNickname, targetHandle, noteID, inReplyTo string) error {
	targetActor := remoteActorURLFromHandle(targetHandle)
	if targetActor == "" {
		return fmt.Errorf("invalid target handle")
	}
	if noteID == "" {
		return fmt.Errorf("note id required")
	}
	object := map[string]any{
		"id":   noteID,
		"type": "Note",
	}
	if inReplyTo != "" {
		object["inReplyTo"] = inReplyTo
	}
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       fmt.Sprintf("%s/deletes/%s", actorURL(actorNickname), uuid.NewString()),
		"type":     "Delete",
		"actor":    actorURL(actorNickname),
		"object":   object,
		"to":       []string{"https://www.w3.org/ns/activitystreams#Public", targetActor},
	}
	inbox := strings.TrimSuffix(targetActor, "/") + "/inbox"
	if d.blobs != nil {
		if resolved := d.resolveInboxURL(targetActor); resolved != "" {
			inbox = resolved
		}
	}
	return d.postActivity(inbox, activity)
}

func (d *Delivery) ResolveRemote(parsed social.ParsedHandle) (*social.UserSearchResult, error) {
	if !config.Cfg.Federation.Enabled {
		return nil, social.ErrRemoteNotReady
	}
	actor, err := fetchActor(d.client, d.blobs, parsed)
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

func fetchActor(client *http.Client, blobs blob.Store, parsed social.ParsedHandle) (map[string]any, error) {
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
	var actorURLStr string
	for _, link := range links {
		m, _ := link.(map[string]any)
		if m["rel"] == "self" {
			actorURLStr, _ = m["href"].(string)
			break
		}
	}
	if actorURLStr == "" {
		return nil, social.ErrUserNotFound
	}

	actorReq, err := http.NewRequest(http.MethodGet, actorURLStr, nil)
	if err != nil {
		return nil, err
	}
	actorReq.Header.Set("Accept", "application/activity+json")
	if err := signOutboundGET(blobs, actorReq); err != nil {
		slog.Debug("federation actor fetch unsigned", "url", actorURLStr, "err", err)
	}
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

func signOutboundGET(blobs blob.Store, req *http.Request) error {
	if blobs == nil {
		return fmt.Errorf("no blobs")
	}
	ak, err := LoadOrCreateInstanceActorKey(blobs)
	if err != nil {
		return err
	}
	return httpsig.SignGET(req, ak.Private, ak.KeyID)
}

func signOutboundPOST(blobs blob.Store, actorNickname string, req *http.Request, body []byte) error {
	if blobs == nil {
		return fmt.Errorf("no blobs")
	}
	ak, err := LoadOrCreateUserActorKey(blobs, actorNickname)
	if err != nil {
		return err
	}
	return httpsig.SignPOST(req, body, ak.Private, ak.KeyID)
}

func buildWorkoutObject(authorNickname string, workout *workouts.Workout, trackData []byte, mediaFiles []workouts.MediaFileInput) map[string]any {
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
		"mediaItems":      []map[string]any{},
	}
	if workout.DurationTotalSeconds > 0 {
		object["durationTotalSeconds"] = workout.DurationTotalSeconds
	}
	if workout.TempAvgKmm != nil && *workout.TempAvgKmm != "" {
		object["tempAvgKmm"] = *workout.TempAvgKmm
	}
	if workout.ElevationGain != nil && *workout.ElevationGain > 0 {
		object["elevationGain"] = *workout.ElevationGain
	}
	if workout.SpeedAvgKmh != nil && *workout.SpeedAvgKmh > 0 {
		object["speedAvgKmh"] = *workout.SpeedAvgKmh
	}
	if workout.HeartRateAvg != nil && *workout.HeartRateAvg > 0 {
		object["heartRateAvg"] = *workout.HeartRateAvg
	}
	if workout.StepsTotal != nil && *workout.StepsTotal > 0 {
		object["stepsTotal"] = *workout.StepsTotal
	}
	if workout.Calories != nil && *workout.Calories > 0 {
		object["calories"] = *workout.Calories
	}
	if workout.LikesCount > 0 {
		object["likesCount"] = workout.LikesCount
	}
	if len(workout.LikedUsers) > 0 {
		users := make([]map[string]any, 0, len(workout.LikedUsers))
		for _, user := range workout.LikedUsers {
			users = append(users, map[string]any{
				"handle":    user.Handle,
				"nickname":  user.Nickname,
				"name":      user.Name,
				"is_local":  HandleIsLocal(user.Handle),
				"avatarUrl": ExportLikeUserAvatarURL(user),
			})
		}
		object["likedUsers"] = users
	}
	if workout.CommentsCount > 0 {
		object["commentsCount"] = workout.CommentsCount
	}
	if len(workout.Comments) > 0 {
		comments := make([]map[string]any, 0, len(workout.Comments))
		for _, c := range workout.Comments {
			entry := map[string]any{
				"id":       c.ID,
				"datetime": c.Datetime.UTC().Format(time.RFC3339),
				"text":     c.Text,
				"user": map[string]any{
					"handle":    c.User.Handle,
					"nickname":  c.User.Nickname,
					"name":      c.User.Name,
					"is_local":  HandleIsLocal(c.User.Handle),
					"avatarUrl": ExportLikeUserAvatarURL(c.User),
				},
			}
			if c.NoteID != "" {
				entry["noteId"] = c.NoteID
			}
			comments = append(comments, entry)
		}
		object["comments"] = comments
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
	return object
}

func (d *Delivery) deliverWorkoutActivity(activityType, authorNickname string, workout *workouts.Workout, followerInboxes []string, trackData []byte, mediaFiles []workouts.MediaFileInput) error {
	if len(followerInboxes) == 0 {
		return nil
	}
	author := actorURL(authorNickname)
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       fmt.Sprintf("%s/activities/%s", author, uuid.NewString()),
		"type":     activityType,
		"actor":    author,
		"object":   buildWorkoutObject(authorNickname, workout, trackData, mediaFiles),
		"to":       []string{"https://www.w3.org/ns/activitystreams#Public"},
	}
	for _, inbox := range DeduplicateInboxURLs(followerInboxes) {
		if err := d.postActivity(inbox, activity); err != nil {
			slog.Error("federation workout activity delivery failed",
				"activity_type", activityType,
				"author", authorNickname,
				"workout_id", workout.ID,
				"inbox", inbox,
				"err", err,
			)
			return err
		}
	}
	slog.Info("federation workout activity delivered",
		"activity_type", activityType,
		"author", authorNickname,
		"workout_id", workout.ID,
		"inboxes", len(followerInboxes),
	)
	return nil
}

func (d *Delivery) DeliverWorkout(authorNickname string, workout *workouts.Workout, followerInboxes []string, trackData []byte, mediaFiles []workouts.MediaFileInput) error {
	return d.deliverWorkoutActivity("Create", authorNickname, workout, followerInboxes, trackData, mediaFiles)
}

func (d *Delivery) DeliverWorkoutUpdate(authorNickname string, workout *workouts.Workout, followerInboxes []string, trackData []byte, mediaFiles []workouts.MediaFileInput) error {
	return d.deliverWorkoutActivity("Update", authorNickname, workout, followerInboxes, trackData, mediaFiles)
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
	for _, inbox := range DeduplicateInboxURLs(followerInboxes) {
		if err := d.postActivity(inbox, activity); err != nil {
			slog.Error("federation workout delete delivery failed",
				"author", authorNickname,
				"workout_id", workoutID,
				"inbox", inbox,
				"err", err,
			)
			return err
		}
	}
	slog.Info("federation workout delete delivered",
		"author", authorNickname,
		"workout_id", workoutID,
		"inboxes", len(followerInboxes),
	)
	return nil
}

// DeliverActorDelete notifies remote inboxes that the local actor was deleted.
// Delivery is best-effort: individual inbox failures are logged and do not stop the loop.
func (d *Delivery) DeliverActorDelete(nickname string, followerInboxes []string) {
	if d == nil || nickname == "" || len(followerInboxes) == 0 {
		return
	}
	author := actorURL(nickname)
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       fmt.Sprintf("%s/activities/%s", author, uuid.NewString()),
		"type":     "Delete",
		"actor":    author,
		"object":   author,
		"to":       []string{"https://www.w3.org/ns/activitystreams#Public"},
	}
	for _, inbox := range DeduplicateInboxURLs(followerInboxes) {
		if err := d.postActivity(inbox, activity); err != nil {
			slog.Error("federation actor delete delivery failed",
				"nickname", nickname,
				"inbox", inbox,
				"err", err,
			)
			continue
		}
	}
	slog.Info("federation actor delete delivered",
		"nickname", nickname,
		"inboxes", len(followerInboxes),
	)
}
