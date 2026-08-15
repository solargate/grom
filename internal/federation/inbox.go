package federation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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
	users             users.Repository
	social            *social.Service
	delivery          *Delivery
	inboxStore        InboxRepository
	likes             workouts.LikesRepository
	comments          workouts.CommentsRepository
	followersStore    FollowersRepository
	onWorkoutLike     func(ownerNickname, workoutID string)
	onWorkoutComment  func(ownerNickname, workoutID string)
	autoAccept        bool
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

func (p *InboxProcessor) SetLikes(likes workouts.LikesRepository, onWorkoutLike func(ownerNickname, workoutID string)) {
	p.likes = likes
	p.onWorkoutLike = onWorkoutLike
}

func (p *InboxProcessor) SetComments(comments workouts.CommentsRepository, onWorkoutComment func(ownerNickname, workoutID string)) {
	p.comments = comments
	p.onWorkoutComment = onWorkoutComment
}

func (p *InboxProcessor) Handle(nickname string, body io.Reader) error {
	var activity map[string]any
	if err := json.NewDecoder(body).Decode(&activity); err != nil {
		slog.Warn("federation inbox invalid JSON", "nickname", nickname, "err", err)
		return err
	}

	actType, _ := activity["type"].(string)
	var err error
	switch actType {
	case "Follow":
		err = p.handleFollow(nickname, activity)
	case "Accept":
		err = p.handleAccept(nickname, activity)
	case "Create":
		err = p.handleCreate(nickname, activity)
	case "Update":
		err = p.handleUpdate(nickname, activity)
	case "Delete":
		err = p.handleDelete(nickname, activity)
	case "Like":
		err = p.handleLike(nickname, activity)
	case "Undo":
		err = p.handleUndo(nickname, activity)
	default:
		slog.Debug("federation inbox ignored activity type", "nickname", nickname, "activity_type", actType)
		return nil
	}
	if err != nil {
		slog.Error("federation inbox activity failed",
			"nickname", nickname,
			"activity_type", actType,
			"err", err,
		)
		return err
	}
	slog.Info("federation inbox activity processed",
		"nickname", nickname,
		"activity_type", actType,
	)
	return nil
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
	object, ok := activity["object"].(map[string]any)
	if !ok {
		return nil
	}
	if object["type"] == "Like" {
		return p.handleUndoLike(targetNickname, activity, object)
	}
	if p.followersStore == nil || object["type"] != "Follow" {
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

func (p *InboxProcessor) handleUndoLike(targetNickname string, activity, object map[string]any) error {
	if p.likes == nil {
		return nil
	}
	actorURI, _ := activity["actor"].(string)
	handle := actorToHandle(actorURI)
	if handle == "" {
		handle, _ = object["actor"].(string)
		handle = actorToHandle(handle)
	}
	if handle == "" {
		return nil
	}
	workoutID := workoutIDFromDeleteObject(object["object"])
	if workoutID == "" {
		workoutID = workoutIDFromDeleteObject(object)
	}
	if workoutID == "" {
		return nil
	}
	likes, err := p.likes.GetLocal(targetNickname, workoutID)
	if err != nil {
		return err
	}
	updated := workouts.RemoveWorkoutLikeUser(likes, handle)
	if err := p.likes.PutLocal(targetNickname, workoutID, &updated); err != nil {
		return err
	}
	if p.onWorkoutLike != nil {
		p.onWorkoutLike(targetNickname, workoutID)
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
	if object, ok := activity["object"].(map[string]any); ok {
		if typ, _ := object["type"].(string); typ == "Note" && stringValue(object, "inReplyTo") != "" {
			return p.handleCommentCreate(viewerNickname, activity, object)
		}
	}
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
		if err := p.inboxStore.Replace(viewerNickname, ownerHandle, workout, trackData, mediaFiles, actorDoc); err != nil {
			return err
		}
	} else {
		if err := p.inboxStore.Save(viewerNickname, ownerHandle, workout, trackData, mediaFiles, actorDoc); err != nil {
			return err
		}
	}
	if p.likes != nil {
		likes := workouts.NormalizeWorkoutLikes(&workouts.WorkoutLikes{
			Likes: workout.LikesCount,
			Users: workout.LikedUsers,
		})
		p.ensureRemoteInteractionAvatars(viewerNickname, likes.Users)
		if likes.Likes == 0 && len(likes.Users) == 0 {
			_ = p.likes.DeleteFederated(viewerNickname, ownerHandle, workout.ID)
		} else if err := p.likes.PutFederated(viewerNickname, ownerHandle, workout.ID, &likes); err != nil {
			return err
		}
	}
	if p.comments != nil {
		comments := workouts.NormalizeWorkoutComments(&workouts.WorkoutComments{
			CommentsNum: workout.CommentsCount,
			Comments:    workout.Comments,
		})
		commentUsers := make([]workouts.WorkoutLikeUser, 0, len(comments.Comments))
		for _, c := range comments.Comments {
			commentUsers = append(commentUsers, c.User)
		}
		p.ensureRemoteInteractionAvatars(viewerNickname, commentUsers)
		if comments.CommentsNum == 0 {
			_ = p.comments.DeleteFederated(viewerNickname, ownerHandle, workout.ID)
		} else if err := p.comments.PutFederated(viewerNickname, ownerHandle, workout.ID, &comments); err != nil {
			return err
		}
	}
	return nil
}

func (p *InboxProcessor) ensureRemoteInteractionAvatars(viewerNickname string, authors []workouts.WorkoutLikeUser) {
	if p.inboxStore == nil || viewerNickname == "" {
		return
	}
	for _, u := range authors {
		if u.Handle == "" || HandleIsLocal(u.Handle) {
			continue
		}
		avatarURL := u.AvatarURL
		if avatarURL == "" {
			avatarURL = publicAvatarURLForHandle(u.Handle, u.Nickname)
		}
		if avatarURL == "" {
			continue
		}
		_ = p.inboxStore.EnsureAuthor(viewerNickname, u.Handle, u.Nickname, u.Name, avatarURL, false)
	}
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
	if count, ok := optionalPositiveInt(object, "likesCount"); ok {
		workout.LikesCount = count
	}
	if likedUsers, ok := object["likedUsers"].([]any); ok {
		workout.LikedUsers = parseFederatedLikedUsers(likedUsers)
	}
	if count, ok := optionalPositiveInt(object, "commentsCount"); ok {
		workout.CommentsCount = count
	}
	if comments, ok := object["comments"].([]any); ok {
		workout.Comments = parseFederatedComments(comments)
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

func parseFederatedLikedUsers(items []any) []workouts.WorkoutLikeUser {
	users := make([]workouts.WorkoutLikeUser, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		users = append(users, parseFederatedLikeUser(obj))
	}
	return workouts.NormalizeWorkoutLikes(&workouts.WorkoutLikes{Users: users}).Users
}

func parseFederatedComments(items []any) []workouts.WorkoutComment {
	comments := make([]workouts.WorkoutComment, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		userMap, _ := m["user"].(map[string]any)
		user := workouts.WorkoutLikeUser{}
		if userMap != nil {
			user = parseFederatedLikeUser(userMap)
		}
		var dt time.Time
		if raw := stringValue(m, "datetime"); raw != "" {
			if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
				dt = parsed
			}
		}
		comments = append(comments, workouts.WorkoutComment{
			ID:       stringValue(m, "id"),
			User:     user,
			Datetime: dt,
			Text:     stringValue(m, "text"),
			NoteID:   firstString(m, "noteId", "note_id"),
		})
	}
	return workouts.NormalizeWorkoutComments(&workouts.WorkoutComments{Comments: comments}).Comments
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v := stringValue(m, k); v != "" {
			return v
		}
	}
	return ""
}

func (p *InboxProcessor) handleLike(targetNickname string, activity map[string]any) error {
	if p.likes == nil {
		return nil
	}
	actorURI, _ := activity["actor"].(string)
	handle := actorToHandle(actorURI)
	if handle == "" {
		return nil
	}
	workoutID := workoutIDFromDeleteObject(activity["object"])
	if workoutID == "" {
		return nil
	}
	actorUser := workouts.WorkoutLikeUser{
		Handle:   handle,
		Nickname: OwnerNicknameFromKey(OwnerKeyFromHandle(handle)),
		Name:     OwnerNicknameFromKey(OwnerKeyFromHandle(handle)),
		IsLocal:  false,
	}
	if p.delivery != nil {
		parsed := social.ParsedHandle{
			Nickname: actorUser.Nickname,
			Domain:   domainFromHandle(handle),
			Handle:   handle,
		}
		if fetched, err := fetchActor(p.delivery.Client(), parsed); err == nil {
			if name := ExtractActorName(fetched); name != "" {
				actorUser.Name = name
			}
			if icon := ExtractIconURL(fetched); icon != "" {
				actorUser.AvatarURL = icon
			}
		}
	}
	if actorUser.AvatarURL == "" {
		actorUser.AvatarURL = publicAvatarURLForHandle(handle, actorUser.Nickname)
	}
	if p.inboxStore != nil && actorUser.AvatarURL != "" {
		_ = p.inboxStore.EnsureAuthor(targetNickname, handle, actorUser.Nickname, actorUser.Name, actorUser.AvatarURL, true)
	}
	likes, err := p.likes.GetLocal(targetNickname, workoutID)
	if err != nil {
		return err
	}
	updated := workouts.AddWorkoutLikeUser(likes, actorUser)
	if err := p.likes.PutLocal(targetNickname, workoutID, &updated); err != nil {
		return err
	}
	if p.onWorkoutLike != nil {
		p.onWorkoutLike(targetNickname, workoutID)
	}
	return nil
}

func (p *InboxProcessor) handleCommentCreate(targetNickname string, activity, object map[string]any) error {
	if p.comments == nil {
		return nil
	}
	inReplyTo := stringValue(object, "inReplyTo")
	workoutID := workoutIDFromObject(map[string]any{"id": inReplyTo})
	if workoutID == "" {
		return nil
	}
	noteID := stringValue(object, "id")
	if noteID == "" {
		return nil
	}
	text := strings.TrimSpace(stringValue(object, "content"))
	if text == "" {
		return nil
	}
	actorURI, _ := activity["actor"].(string)
	handle := actorToHandle(actorURI)
	if handle == "" {
		return nil
	}
	actorUser := workouts.WorkoutLikeUser{
		Handle:   handle,
		Nickname: OwnerNicknameFromKey(OwnerKeyFromHandle(handle)),
		Name:     OwnerNicknameFromKey(OwnerKeyFromHandle(handle)),
		IsLocal:  false,
	}
	if p.delivery != nil {
		parsed := social.ParsedHandle{
			Nickname: actorUser.Nickname,
			Domain:   domainFromHandle(handle),
			Handle:   handle,
		}
		if fetched, err := fetchActor(p.delivery.Client(), parsed); err == nil {
			if name := ExtractActorName(fetched); name != "" {
				actorUser.Name = name
			}
			if icon := ExtractIconURL(fetched); icon != "" {
				actorUser.AvatarURL = icon
			}
		}
	}
	if actorUser.AvatarURL == "" {
		actorUser.AvatarURL = publicAvatarURLForHandle(handle, actorUser.Nickname)
	}
	if p.inboxStore != nil && actorUser.AvatarURL != "" {
		_ = p.inboxStore.EnsureAuthor(targetNickname, handle, actorUser.Nickname, actorUser.Name, actorUser.AvatarURL, true)
	}
	var dt time.Time
	if raw := stringValue(object, "published"); raw != "" {
		if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
			dt = parsed
		}
	}
	if dt.IsZero() {
		dt = time.Now().UTC()
	}
	// Prefer Note id suffix as stable comment id when it looks like a UUID path.
	commentID := noteID
	if idx := strings.LastIndex(noteID, "/notes/"); idx >= 0 {
		commentID = noteID[idx+len("/notes/"):]
	}
	if commentID == "" {
		commentID = uuid.NewString()
	}
	comments, err := p.comments.GetLocal(targetNickname, workoutID)
	if err != nil {
		return err
	}
	updated := workouts.AddWorkoutComment(comments, workouts.WorkoutComment{
		ID:       commentID,
		User:     actorUser,
		Datetime: dt,
		Text:     text,
		NoteID:   noteID,
	})
	if err := p.comments.PutLocal(targetNickname, workoutID, &updated); err != nil {
		return err
	}
	if p.onWorkoutComment != nil {
		p.onWorkoutComment(targetNickname, workoutID)
	}
	return nil
}

func (p *InboxProcessor) handleDelete(viewerNickname string, activity map[string]any) error {
	if noteID, inReplyTo, ok := noteDeleteTarget(activity["object"]); ok {
		return p.handleCommentDelete(viewerNickname, activity, noteID, inReplyTo)
	}

	actorURI, _ := activity["actor"].(string)
	ownerHandle := actorToHandle(actorURI)
	if ownerHandle == "" {
		return nil
	}

	if isActorDelete(activity["object"], actorURI) {
		return p.handleActorDelete(viewerNickname, ownerHandle, actorURI)
	}

	if p.inboxStore == nil {
		return nil
	}

	workoutID := workoutIDFromDeleteObject(activity["object"])
	if workoutID == "" {
		return nil
	}
	if err := p.inboxStore.Delete(viewerNickname, ownerHandle, workoutID); err != nil {
		return err
	}
	if p.likes != nil {
		_ = p.likes.DeleteFederated(viewerNickname, ownerHandle, workoutID)
	}
	if p.comments != nil {
		_ = p.comments.DeleteFederated(viewerNickname, ownerHandle, workoutID)
	}
	return nil
}

func isActorDelete(object any, actorURI string) bool {
	switch obj := object.(type) {
	case string:
		return obj != "" && (obj == actorURI || strings.EqualFold(obj, actorURI))
	case map[string]any:
		id := stringValue(obj, "id")
		typ, _ := obj["type"].(string)
		if id == "" {
			return false
		}
		if typ == "Person" || typ == "Tombstone" || typ == "" {
			return id == actorURI || strings.EqualFold(id, actorURI)
		}
	}
	return false
}

func (p *InboxProcessor) handleActorDelete(viewerNickname, ownerHandle, actorURI string) error {
	if p.inboxStore != nil {
		feed, err := p.inboxStore.List(viewerNickname)
		if err != nil {
			return err
		}
		for _, item := range feed {
			if !strings.EqualFold(item.Author.Handle, ownerHandle) {
				continue
			}
			if p.likes != nil {
				_ = p.likes.DeleteFederated(viewerNickname, ownerHandle, item.ID)
			}
			if p.comments != nil {
				_ = p.comments.DeleteFederated(viewerNickname, ownerHandle, item.ID)
			}
		}
		if err := p.inboxStore.DeleteAllForOwner(viewerNickname, ownerHandle); err != nil {
			return err
		}
	}
	if p.followersStore != nil && actorURI != "" {
		_ = p.followersStore.Remove(viewerNickname, actorURI)
	}
	if p.social != nil && p.users != nil {
		if user, err := p.users.FindByNickname(viewerNickname); err == nil {
			_ = p.social.DeleteFollowsToTarget(user.ID, ownerHandle)
		}
	}
	return nil
}

func noteDeleteTarget(raw any) (noteID, inReplyTo string, ok bool) {
	switch object := raw.(type) {
	case string:
		if strings.Contains(object, "/notes/") {
			return object, "", true
		}
	case map[string]any:
		typ, _ := object["type"].(string)
		id := stringValue(object, "id")
		if typ == "Note" || strings.Contains(id, "/notes/") {
			return id, stringValue(object, "inReplyTo"), id != ""
		}
	}
	return "", "", false
}

func (p *InboxProcessor) handleCommentDelete(viewerNickname string, activity map[string]any, noteID, inReplyTo string) error {
	if p.comments == nil || noteID == "" {
		return nil
	}
	workoutID := workoutIDFromObject(map[string]any{"id": inReplyTo})
	if workoutID == "" {
		// Local owner path: try removing from local comments if workout id unknown —
		// only when this actor targeted our local store via prior Create.
		return nil
	}

	// Prefer local workout comments (we are the workout owner).
	if comments, err := p.comments.GetLocal(viewerNickname, workoutID); err == nil {
		if workouts.FindCommentByNoteID(comments, noteID) != nil {
			updated := workouts.RemoveWorkoutCommentByNoteID(comments, noteID)
			if err := p.comments.PutLocal(viewerNickname, workoutID, &updated); err != nil {
				return err
			}
			if p.onWorkoutComment != nil {
				p.onWorkoutComment(viewerNickname, workoutID)
			}
			return nil
		}
	}

	// Federated cache (viewer had commented / received Update snapshot).
	// Prefer workout owner from inReplyTo: Delete actor may be the comment author,
	// not the workout owner (and owner may delete someone else's comment).
	ownerHandle := ""
	if inReplyTo != "" {
		ownerHandle = ownerHandleFromWorkoutURL(inReplyTo)
	}
	if ownerHandle == "" {
		actorURI, _ := activity["actor"].(string)
		ownerHandle = actorToHandle(actorURI)
	}
	if ownerHandle == "" {
		return nil
	}
	cached, err := p.comments.GetFederated(viewerNickname, ownerHandle, workoutID)
	if err != nil {
		return err
	}
	updated := workouts.RemoveWorkoutCommentByNoteID(cached, noteID)
	if updated.CommentsNum == 0 {
		return p.comments.DeleteFederated(viewerNickname, ownerHandle, workoutID)
	}
	return p.comments.PutFederated(viewerNickname, ownerHandle, workoutID, &updated)
}

func ownerHandleFromWorkoutURL(workoutURL string) string {
	// https://domain/users/nick/workouts/id → nick@domain
	workoutURL = strings.TrimSuffix(workoutURL, "/")
	if !strings.HasPrefix(workoutURL, "https://") {
		return ""
	}
	rest := strings.TrimPrefix(workoutURL, "https://")
	parts := strings.Split(rest, "/")
	if len(parts) < 4 || parts[1] != "users" {
		return ""
	}
	domain := parts[0]
	nick := parts[2]
	if domain == "" || nick == "" {
		return ""
	}
	return nick + "@" + domain
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
