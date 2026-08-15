package bbolt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/maprender"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

type InboxStore struct {
	db              *bolt.DB
	blobs           blob.Store
	speedCharts     workouts.SpeedChartStore
	heartRateCharts workouts.HeartRateChartStore
	client          *http.Client
}

func NewInboxStore(db *bolt.DB, blobs blob.Store, speedCharts workouts.SpeedChartStore, heartRateCharts workouts.HeartRateChartStore) *InboxStore {
	return &InboxStore{db: db, blobs: blobs, speedCharts: speedCharts, heartRateCharts: heartRateCharts}
}

func (s *InboxStore) SetHTTPClient(client *http.Client) {
	s.client = client
}

func fedAuthorKey(viewer, ownerKey string) []byte {
	return []byte(viewer + "/" + ownerKey)
}

func fedInboxKey(viewer, ownerKey, workoutID string) []byte {
	return []byte(viewer + "/" + ownerKey + "/" + workoutID)
}

func fedInboxViewerPrefix(viewer string) []byte {
	return []byte(viewer + "/")
}

func (s *InboxStore) getAuthor(tx *bolt.Tx, viewer, ownerKey string) (federation.AuthorMeta, error) {
	raw := tx.Bucket(bucketFedAuthors).Get(fedAuthorKey(viewer, ownerKey))
	if raw == nil {
		return federation.AuthorMeta{}, nil
	}
	var meta federation.AuthorMeta
	if err := unmarshalJSON(raw, &meta); err != nil {
		return federation.AuthorMeta{}, err
	}
	return meta, nil
}

func (s *InboxStore) putAuthor(tx *bolt.Tx, viewer, ownerKey string, meta federation.AuthorMeta) error {
	raw, err := marshalJSON(meta)
	if err != nil {
		return err
	}
	return tx.Bucket(bucketFedAuthors).Put(fedAuthorKey(viewer, ownerKey), raw)
}

func (s *InboxStore) getWorkout(tx *bolt.Tx, viewer, ownerKey, workoutID string) (*workouts.Workout, error) {
	raw := tx.Bucket(bucketFedInbox).Get(fedInboxKey(viewer, ownerKey, workoutID))
	if raw == nil {
		return nil, workouts.ErrWorkoutNotFound
	}
	var w workouts.Workout
	if err := unmarshalJSON(raw, &w); err != nil {
		return nil, fmt.Errorf("parse federated workout: %w", err)
	}
	ctx := context.Background()
	if exists, err := s.blobs.Exists(ctx, keys.FederatedInboxMapPreview(viewer, ownerKey, workoutID)); err == nil && exists {
		w.HasMapPreview = true
	}
	w.HasMedia = len(w.MediaFiles) > 0
	return &w, nil
}

func (s *InboxStore) putWorkout(tx *bolt.Tx, viewer, ownerKey string, w *workouts.Workout) error {
	raw, err := marshalJSON(w)
	if err != nil {
		return err
	}
	return tx.Bucket(bucketFedInbox).Put(fedInboxKey(viewer, ownerKey, w.ID), raw)
}

func (s *InboxStore) findOwnerKey(tx *bolt.Tx, viewer, ownerNickname string) (string, error) {
	prefix := fedInboxViewerPrefix(viewer)
	seen := make(map[string]struct{})
	c := tx.Bucket(bucketFedInbox).Cursor()
	for k, _ := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, _ = c.Next() {
		parts := strings.SplitN(string(k), "/", 3)
		if len(parts) != 3 {
			continue
		}
		ownerKey := parts[1]
		if _, ok := seen[ownerKey]; ok {
			continue
		}
		seen[ownerKey] = struct{}{}
		if strings.EqualFold(federation.OwnerNicknameFromKey(ownerKey), ownerNickname) {
			return ownerKey, nil
		}
	}
	// Also check authors-only (no workouts yet).
	ac := tx.Bucket(bucketFedAuthors).Cursor()
	for k, _ := ac.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, _ = ac.Next() {
		parts := strings.SplitN(string(k), "/", 2)
		if len(parts) != 2 {
			continue
		}
		ownerKey := parts[1]
		if strings.EqualFold(federation.OwnerNicknameFromKey(ownerKey), ownerNickname) {
			return ownerKey, nil
		}
	}
	return "", workouts.ErrWorkoutNotFound
}

func (s *InboxStore) mergeAuthor(viewer, ownerHandle, nickname string, actor map[string]any, refresh bool) error {
	ownerKey := federation.OwnerKeyFromHandle(ownerHandle)
	return s.db.Update(func(tx *bolt.Tx) error {
		meta, err := s.getAuthor(tx, viewer, ownerKey)
		if err != nil {
			return err
		}
		federation.UpdateAuthorMeta(&meta, ownerHandle, nickname, actor, s.client, refresh, s.blobs, viewer, ownerKey)
		if meta.Handle == "" && meta.Nickname == "" {
			return nil
		}
		return s.putAuthor(tx, viewer, ownerKey, meta)
	})
}

func (s *InboxStore) EnsureAuthor(viewerNickname, handle, nickname, name, remoteAvatarURL string, refresh bool) error {
	actor := map[string]any{}
	if name != "" {
		actor["name"] = name
	}
	if remoteAvatarURL != "" {
		actor["icon"] = map[string]any{"url": remoteAvatarURL}
	}
	if len(actor) == 0 {
		actor = nil
	}
	return s.mergeAuthor(viewerNickname, handle, nickname, actor, refresh)
}

func (s *InboxStore) AuthorAvatarFields(viewerNickname, handle string) (bool, string) {
	ownerKey := federation.OwnerKeyFromHandle(handle)
	var meta federation.AuthorMeta
	_ = s.db.View(func(tx *bolt.Tx) error {
		var err error
		meta, err = s.getAuthor(tx, viewerNickname, ownerKey)
		return err
	})
	return federation.AuthorAvatarFields(s.blobs, viewerNickname, ownerKey, handle, meta)
}

func (s *InboxStore) writeFederatedTrack(ctx context.Context, viewerNickname, ownerKey string, workout *workouts.Workout, trackData []byte) error {
	trackKey := keys.FederatedInboxTrack(viewerNickname, ownerKey, workout.ID, workout.Track)
	if _, err := blob.PutBytes(ctx, s.blobs, trackKey, trackData, blob.PutOptions{}); err != nil {
		return err
	}

	parsed, err := tracks.Parse(trackData, workout.Track)
	if err != nil {
		slog.Warn("federated track parse failed", "workout_id", workout.ID, "err", err)
		return nil
	}

	if s.speedCharts != nil {
		if err := s.speedCharts.WriteFederated(ctx, viewerNickname, ownerKey, workout.ID, workouts.BuildSpeedChartSamples(parsed)); err != nil {
			slog.Error("federated speed chart write failed", "workout_id", workout.ID, "err", err)
		}
	}

	if s.heartRateCharts != nil {
		if err := s.heartRateCharts.WriteFederated(ctx, viewerNickname, ownerKey, workout.ID, workouts.BuildHeartRateChartSamples(parsed)); err != nil {
			slog.Error("federated heart rate chart write failed", "workout_id", workout.ID, "err", err)
		}
	}

	if parsed.HasGPS() {
		if preview, err := maprender.RenderPreview(parsed.Points); err != nil {
			slog.Error("federated map preview render failed", "workout_id", workout.ID, "err", err)
		} else if len(preview) > 0 {
			previewKey := keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workout.ID)
			if _, err := blob.PutBytes(ctx, s.blobs, previewKey, preview, blob.PutOptions{ContentType: "image/webp"}); err != nil {
				slog.Error("federated map preview write failed", "workout_id", workout.ID, "err", err)
			} else {
				workout.HasMapPreview = true
			}
		}
	}
	return nil
}

func (s *InboxStore) Save(viewerNickname, ownerHandle string, workout *workouts.Workout, trackData []byte, mediaFiles []workouts.MediaFileInput, actor map[string]any) error {
	ownerKey := federation.OwnerKeyFromHandle(ownerHandle)
	nickname := federation.OwnerNicknameFromKey(ownerKey)
	if err := s.mergeAuthor(viewerNickname, ownerHandle, nickname, actor, actor != nil && federation.ExtractIconURL(actor) != ""); err != nil {
		return err
	}

	ctx := context.Background()
	if workout.Track != "" && len(trackData) > 0 {
		if err := s.writeFederatedTrack(ctx, viewerNickname, ownerKey, workout, trackData); err != nil {
			return err
		}
	}
	if len(mediaFiles) > 0 {
		savedNames, err := workouts.SaveFederatedMedia(s.blobs, viewerNickname, ownerKey, workout.ID, mediaFiles)
		if err != nil {
			return err
		}
		workout.MediaFiles = savedNames
		workout.HasMedia = len(savedNames) > 0
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		return s.putWorkout(tx, viewerNickname, ownerKey, workout)
	})
}

func (s *InboxStore) Replace(viewerNickname, ownerHandle string, workout *workouts.Workout, trackData []byte, mediaFiles []workouts.MediaFileInput, actor map[string]any) error {
	ownerKey := federation.OwnerKeyFromHandle(ownerHandle)
	nickname := federation.OwnerNicknameFromKey(ownerKey)
	if err := s.mergeAuthor(viewerNickname, ownerHandle, nickname, actor, actor != nil && federation.ExtractIconURL(actor) != ""); err != nil {
		return err
	}

	ctx := context.Background()
	var previous *workouts.Workout
	_ = s.db.View(func(tx *bolt.Tx) error {
		previous, _ = s.getWorkout(tx, viewerNickname, ownerKey, workout.ID)
		return nil
	})

	if workout.Track != "" && len(trackData) > 0 {
		if previous != nil && previous.Track != "" && previous.Track != workout.Track {
			_ = s.blobs.Delete(ctx, keys.FederatedInboxTrack(viewerNickname, ownerKey, workout.ID, previous.Track))
		}
		if err := s.writeFederatedTrack(ctx, viewerNickname, ownerKey, workout, trackData); err != nil {
			return err
		}
	} else if workout.Track == "" {
		if previous != nil && previous.Track != "" {
			_ = s.blobs.Delete(ctx, keys.FederatedInboxTrack(viewerNickname, ownerKey, workout.ID, previous.Track))
		}
		_ = s.blobs.Delete(ctx, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workout.ID))
		if s.speedCharts != nil {
			_ = s.speedCharts.DeleteFederated(ctx, viewerNickname, ownerKey, workout.ID)
		}
		if s.heartRateCharts != nil {
			_ = s.heartRateCharts.DeleteFederated(ctx, viewerNickname, ownerKey, workout.ID)
		}
		workout.HasMapPreview = false
	} else if previous != nil {
		if exists, err := s.blobs.Exists(ctx, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workout.ID)); err == nil && exists {
			workout.HasMapPreview = true
		}
	}

	var previousMedia []string
	if previous != nil {
		previousMedia = previous.MediaFiles
	}
	savedNames, err := workouts.ReplaceFederatedMedia(s.blobs, viewerNickname, ownerKey, workout.ID, previousMedia, mediaFiles)
	if err != nil {
		return err
	}
	workout.MediaFiles = savedNames
	workout.HasMedia = len(savedNames) > 0

	return s.db.Update(func(tx *bolt.Tx) error {
		return s.putWorkout(tx, viewerNickname, ownerKey, workout)
	})
}

func (s *InboxStore) Delete(viewerNickname, ownerHandle, workoutID string) error {
	ownerKey := federation.OwnerKeyFromHandle(ownerHandle)
	var workout *workouts.Workout
	err := s.db.Update(func(tx *bolt.Tx) error {
		w, err := s.getWorkout(tx, viewerNickname, ownerKey, workoutID)
		if err != nil {
			if errors.Is(err, workouts.ErrWorkoutNotFound) {
				return nil
			}
			return err
		}
		workout = w
		if err := DeleteFederatedSpeedChartInTx(tx, viewerNickname, ownerKey, workoutID); err != nil {
			return err
		}
		if err := DeleteFederatedHeartRateChartInTx(tx, viewerNickname, ownerKey, workoutID); err != nil {
			return err
		}
		return tx.Bucket(bucketFedInbox).Delete(fedInboxKey(viewerNickname, ownerKey, workoutID))
	})
	if err != nil || workout == nil {
		return err
	}

	ctx := context.Background()
	if workout.Track != "" {
		_ = s.blobs.Delete(ctx, keys.FederatedInboxTrack(viewerNickname, ownerKey, workoutID, workout.Track))
	}
	_ = s.blobs.Delete(ctx, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workoutID))
	for _, name := range workout.MediaFiles {
		_ = s.blobs.Delete(ctx, keys.FederatedInboxMediaOriginal(viewerNickname, ownerKey, workoutID, name))
		_ = s.blobs.Delete(ctx, keys.FederatedInboxMediaPreview(viewerNickname, ownerKey, workoutID, name))
	}
	return nil
}

func (s *InboxStore) DeleteAllForOwner(viewerNickname, ownerHandle string) error {
	if viewerNickname == "" || ownerHandle == "" {
		return nil
	}
	ownerKey := federation.OwnerKeyFromHandle(ownerHandle)
	prefix := fedInboxKey(viewerNickname, ownerKey, "")
	// fedInboxKey with empty workoutID still ends with trailing slash — collect workout IDs.
	var workoutIDs []string
	err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketFedInbox).Cursor()
		for k, _ := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, _ = c.Next() {
			parts := strings.SplitN(string(k), "/", 3)
			if len(parts) == 3 && parts[2] != "" {
				workoutIDs = append(workoutIDs, parts[2])
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, id := range workoutIDs {
		if err := s.Delete(viewerNickname, ownerHandle, id); err != nil {
			return err
		}
	}
	ctx := context.Background()
	_ = s.blobs.Delete(ctx, keys.FederatedInboxAvatar(viewerNickname, ownerKey))
	return s.db.Update(func(tx *bolt.Tx) error {
		_ = tx.Bucket(bucketFedAuthors).Delete(fedAuthorKey(viewerNickname, ownerKey))
		return nil
	})
}

func (s *InboxStore) withOwnerWorkout(viewerNickname, ownerNickname, workoutID string, fn func(ownerKey string, w *workouts.Workout) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		ownerKey, err := s.findOwnerKey(tx, viewerNickname, ownerNickname)
		if err != nil {
			return err
		}
		w, err := s.getWorkout(tx, viewerNickname, ownerKey, workoutID)
		if err != nil {
			return err
		}
		return fn(ownerKey, w)
	})
}

func (s *InboxStore) TrackFile(viewerNickname, ownerNickname, workoutID string) ([]byte, string, string, error) {
	var data []byte
	var trackName, workoutName string
	err := s.withOwnerWorkout(viewerNickname, ownerNickname, workoutID, func(ownerKey string, w *workouts.Workout) error {
		if w.Track == "" {
			return workouts.ErrWorkoutNotFound
		}
		var err error
		data, err = blob.ReadAll(context.Background(), s.blobs, keys.FederatedInboxTrack(viewerNickname, ownerKey, workoutID, w.Track))
		if err != nil {
			if os.IsNotExist(err) {
				return workouts.ErrWorkoutNotFound
			}
			return fmt.Errorf("read federated track: %w", err)
		}
		trackName, workoutName = w.Track, w.Name
		return nil
	})
	return data, trackName, workoutName, err
}

func (s *InboxStore) MapPreview(viewerNickname, ownerNickname, workoutID string) ([]byte, error) {
	var data []byte
	err := s.withOwnerWorkout(viewerNickname, ownerNickname, workoutID, func(ownerKey string, _ *workouts.Workout) error {
		var err error
		data, err = blob.ReadAll(context.Background(), s.blobs, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workoutID))
		if err != nil {
			if os.IsNotExist(err) {
				return workouts.ErrWorkoutNotFound
			}
			return err
		}
		return nil
	})
	return data, err
}

func (s *InboxStore) MediaOriginal(viewerNickname, ownerNickname, workoutID, filename string) ([]byte, string, error) {
	var data []byte
	var contentType string
	err := s.withOwnerWorkout(viewerNickname, ownerNickname, workoutID, func(ownerKey string, _ *workouts.Workout) error {
		safeName, err := workouts.SanitizeMediaFilename(filename)
		if err != nil {
			return err
		}
		data, err = blob.ReadAll(context.Background(), s.blobs, keys.FederatedInboxMediaOriginal(viewerNickname, ownerKey, workoutID, safeName))
		if err != nil {
			if os.IsNotExist(err) {
				return workouts.ErrPhotoNotFound
			}
			return err
		}
		contentType = workouts.MediaContentType(safeName)
		return nil
	})
	return data, contentType, err
}

func (s *InboxStore) MediaPreview(viewerNickname, ownerNickname, workoutID, filename string) ([]byte, error) {
	var data []byte
	err := s.withOwnerWorkout(viewerNickname, ownerNickname, workoutID, func(ownerKey string, _ *workouts.Workout) error {
		safeName, err := workouts.SanitizeMediaFilename(filename)
		if err != nil {
			return err
		}
		data, err = blob.ReadAll(context.Background(), s.blobs, keys.FederatedInboxMediaPreview(viewerNickname, ownerKey, workoutID, safeName))
		if err != nil {
			if os.IsNotExist(err) {
				return workouts.ErrPhotoNotFound
			}
			return err
		}
		return nil
	})
	return data, err
}

func (s *InboxStore) Avatar(viewerNickname, ownerKey string) ([]byte, error) {
	ctx := context.Background()
	avatarKey := keys.FederatedInboxAvatar(viewerNickname, ownerKey)
	if exists, err := s.blobs.Exists(ctx, avatarKey); err == nil && exists {
		return blob.ReadAll(ctx, s.blobs, avatarKey)
	}

	var meta federation.AuthorMeta
	err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		meta, err = s.getAuthor(tx, viewerNickname, ownerKey)
		return err
	})
	if err != nil {
		return nil, err
	}
	if meta.Handle == "" && meta.Nickname == "" {
		return nil, avatars.ErrAvatarNotFound
	}

	remoteURL := meta.RemoteAvatarURL
	if remoteURL == "" && (strings.HasPrefix(meta.AvatarURL, "http://") || strings.HasPrefix(meta.AvatarURL, "https://")) {
		remoteURL = meta.AvatarURL
	}
	if remoteURL == "" || s.client == nil {
		return nil, avatars.ErrAvatarNotFound
	}

	// Reuse file-store path via UpdateAuthorMeta refresh.
	err = s.db.Update(func(tx *bolt.Tx) error {
		m, err := s.getAuthor(tx, viewerNickname, ownerKey)
		if err != nil {
			return err
		}
		federation.UpdateAuthorMeta(&m, m.Handle, m.Nickname, map[string]any{
			"icon": map[string]any{"url": remoteURL},
		}, s.client, true, s.blobs, viewerNickname, ownerKey)
		return s.putAuthor(tx, viewerNickname, ownerKey, m)
	})
	if err != nil {
		return nil, err
	}
	return blob.ReadAll(ctx, s.blobs, avatarKey)
}

func (s *InboxStore) Get(viewerNickname, ownerNickname, workoutID string) (*workouts.FeedWorkout, error) {
	var result *workouts.FeedWorkout
	var ownerKey string
	err := s.db.View(func(tx *bolt.Tx) error {
		key, err := s.findOwnerKey(tx, viewerNickname, ownerNickname)
		if err != nil {
			return err
		}
		ownerKey = key
		workout, err := s.getWorkout(tx, viewerNickname, ownerKey, workoutID)
		if err != nil {
			return err
		}
		handle := federation.OwnerHandleFromKey(ownerKey)
		nickname := federation.OwnerNicknameFromKey(ownerKey)
		meta, err := s.getAuthor(tx, viewerNickname, ownerKey)
		if err != nil {
			return err
		}
		if meta.Handle != "" {
			handle = meta.Handle
		}
		if meta.Nickname != "" {
			nickname = meta.Nickname
		}
		hasAvatar, avatarURL := federation.AuthorAvatarFields(s.blobs, viewerNickname, ownerKey, handle, meta)
		result = &workouts.FeedWorkout{
			Workout: *workout,
			Owner:   nickname,
			Author: workouts.FeedAuthor{
				Nickname:  nickname,
				Name:      meta.Name,
				Handle:    handle,
				IsLocal:   false,
				HasAvatar: hasAvatar,
				AvatarURL: avatarURL,
			},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *InboxStore) GetSpeedChart(viewerNickname, ownerNickname, workoutID string) (*workouts.Workout, []workouts.SpeedSample, error) {
	var workout *workouts.Workout
	var ownerKey string
	err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		ownerKey, err = s.findOwnerKey(tx, viewerNickname, ownerNickname)
		if err != nil {
			return err
		}
		workout, err = s.getWorkout(tx, viewerNickname, ownerKey, workoutID)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	if workout.Track == "" || s.speedCharts == nil {
		return workout, nil, nil
	}
	samples, err := s.speedCharts.ReadFederated(context.Background(), viewerNickname, ownerKey, workoutID)
	if err != nil {
		return nil, nil, fmt.Errorf("read federated speed chart: %w", err)
	}
	return workout, samples, nil
}

func (s *InboxStore) GetHeartRateChart(viewerNickname, ownerNickname, workoutID string) (*workouts.Workout, []workouts.HeartRateSample, error) {
	var workout *workouts.Workout
	var ownerKey string
	err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		ownerKey, err = s.findOwnerKey(tx, viewerNickname, ownerNickname)
		if err != nil {
			return err
		}
		workout, err = s.getWorkout(tx, viewerNickname, ownerKey, workoutID)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	if workout.Track == "" || s.heartRateCharts == nil {
		return workout, nil, nil
	}
	samples, err := s.heartRateCharts.ReadFederated(context.Background(), viewerNickname, ownerKey, workoutID)
	if err != nil {
		return nil, nil, fmt.Errorf("read federated heart rate chart: %w", err)
	}
	return workout, samples, nil
}

func (s *InboxStore) List(viewerNickname string) ([]workouts.FeedWorkout, error) {
	prefix := fedInboxViewerPrefix(viewerNickname)
	items := make([]workouts.FeedWorkout, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		authors := make(map[string]federation.AuthorMeta)
		ac := tx.Bucket(bucketFedAuthors).Cursor()
		for k, v := ac.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, v = ac.Next() {
			parts := strings.SplitN(string(k), "/", 2)
			if len(parts) != 2 {
				continue
			}
			var meta federation.AuthorMeta
			if err := unmarshalJSON(v, &meta); err != nil {
				return err
			}
			authors[parts[1]] = meta
		}

		c := tx.Bucket(bucketFedInbox).Cursor()
		for k, v := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, v = c.Next() {
			parts := strings.SplitN(string(k), "/", 3)
			if len(parts) != 3 {
				continue
			}
			ownerKey := parts[1]
			var workout workouts.Workout
			if err := unmarshalJSON(v, &workout); err != nil {
				return err
			}
			ctx := context.Background()
			if exists, err := s.blobs.Exists(ctx, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workout.ID)); err == nil && exists {
				workout.HasMapPreview = true
			}
			workout.HasMedia = len(workout.MediaFiles) > 0

			handle := federation.OwnerHandleFromKey(ownerKey)
			nickname := federation.OwnerNicknameFromKey(ownerKey)
			meta := authors[ownerKey]
			if meta.Handle != "" {
				handle = meta.Handle
			}
			if meta.Nickname != "" {
				nickname = meta.Nickname
			}
			hasAvatar, avatarURL := federation.AuthorAvatarFields(s.blobs, viewerNickname, ownerKey, handle, meta)
			items = append(items, workouts.FeedWorkout{
				Workout: workout,
				Owner:   nickname,
				Author: workouts.FeedAuthor{
					Nickname:  nickname,
					Name:      meta.Name,
					Handle:    handle,
					IsLocal:   false,
					HasAvatar: hasAvatar,
					AvatarURL: avatarURL,
				},
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		return workouts.FeedNewer(items[i].StartDate, items[i].ID, items[j].StartDate, items[j].ID)
	})
	return items, nil
}

func (s *InboxStore) ListPage(viewerNickname string, cursor *workouts.Cursor, limit int) ([]workouts.FeedWorkout, bool, error) {
	if limit <= 0 {
		limit = workouts.DefaultPageLimit
	}
	all, err := s.List(viewerNickname)
	if err != nil {
		return nil, false, err
	}
	filtered := make([]workouts.FeedWorkout, 0, limit)
	for i := range all {
		if !workouts.AfterCursor(all[i].StartDate, all[i].ID, cursor) {
			continue
		}
		filtered = append(filtered, all[i])
		if len(filtered) > limit {
			break
		}
	}
	hasMore := len(filtered) > limit
	if hasMore {
		filtered = filtered[:limit]
	}
	return filtered, hasMore, nil
}

// PutExistingAuthor / PutExistingWorkout used by migration.
func (s *InboxStore) PutExistingAuthor(viewer, ownerKey string, meta federation.AuthorMeta) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.putAuthor(tx, viewer, ownerKey, meta)
	})
}

func (s *InboxStore) PutExistingWorkout(viewer, ownerKey string, w *workouts.Workout) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.putWorkout(tx, viewer, ownerKey, w)
	})
}

func (s *InboxStore) ListAllAuthors() (map[string]map[string]federation.AuthorMeta, error) {
	out := make(map[string]map[string]federation.AuthorMeta)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedAuthors).ForEach(func(k, v []byte) error {
			parts := strings.SplitN(string(k), "/", 2)
			if len(parts) != 2 {
				return nil
			}
			var meta federation.AuthorMeta
			if err := unmarshalJSON(v, &meta); err != nil {
				return err
			}
			if out[parts[0]] == nil {
				out[parts[0]] = make(map[string]federation.AuthorMeta)
			}
			out[parts[0]][parts[1]] = meta
			return nil
		})
	})
	return out, err
}

func (s *InboxStore) ListAllWorkouts() (map[string]map[string][]workouts.Workout, error) {
	// viewer -> ownerKey -> workouts
	out := make(map[string]map[string][]workouts.Workout)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedInbox).ForEach(func(k, v []byte) error {
			parts := strings.SplitN(string(k), "/", 3)
			if len(parts) != 3 {
				return nil
			}
			var w workouts.Workout
			if err := unmarshalJSON(v, &w); err != nil {
				return err
			}
			if out[parts[0]] == nil {
				out[parts[0]] = make(map[string][]workouts.Workout)
			}
			out[parts[0]][parts[1]] = append(out[parts[0]][parts[1]], w)
			return nil
		})
	})
	return out, err
}

var _ federation.InboxRepository = (*InboxStore)(nil)
