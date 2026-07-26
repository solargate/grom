package federation

import (
	"net/http"

	"github.com/solargate/grom/internal/workouts"
)

type InboundFollower struct {
	ActorURI string `yaml:"actor_uri" json:"actor_uri"`
	Inbox    string `yaml:"inbox" json:"inbox"`
	Handle   string `yaml:"handle" json:"handle"`
}

type FollowersRepository interface {
	Add(nickname string, follower InboundFollower) error
	List(nickname string) ([]InboundFollower, error)
	ListInboxes(nickname string) ([]string, error)
	Remove(nickname, actorURI string) error
}

type InboxRepository interface {
	SetHTTPClient(client *http.Client)
	EnsureAuthor(viewerNickname, handle, nickname, name, remoteAvatarURL string, refresh bool) error
	AuthorAvatarFields(viewerNickname, handle string) (bool, string)
	Save(viewerNickname, ownerHandle string, workout *workouts.Workout, trackData []byte, mediaFiles []workouts.MediaFileInput, actor map[string]any) error
	// Replace writes a full federated workout snapshot (metadata + optional track + media set).
	// Empty mediaFiles clears previously stored media. Empty track clears previously stored track artifacts.
	Replace(viewerNickname, ownerHandle string, workout *workouts.Workout, trackData []byte, mediaFiles []workouts.MediaFileInput, actor map[string]any) error
	Delete(viewerNickname, ownerHandle, workoutID string) error
	TrackFile(viewerNickname, ownerNickname, workoutID string) ([]byte, string, string, error)
	MapPreview(viewerNickname, ownerNickname, workoutID string) ([]byte, error)
	MediaOriginal(viewerNickname, ownerNickname, workoutID, filename string) ([]byte, string, error)
	MediaPreview(viewerNickname, ownerNickname, workoutID, filename string) ([]byte, error)
	Avatar(viewerNickname, ownerKey string) ([]byte, error)
	Get(viewerNickname, ownerNickname, workoutID string) (*workouts.FeedWorkout, error)
	List(viewerNickname string) ([]workouts.FeedWorkout, error)
	ListPage(viewerNickname string, cursor *workouts.Cursor, limit int) ([]workouts.FeedWorkout, bool, error)
}

type Storage interface {
	Followers() FollowersRepository
	Inbox() InboxRepository
}

var _ InboxRepository = (*WorkoutInboxStore)(nil)

type storage struct {
	followers FollowersRepository
	inbox     InboxRepository
}

func NewStorage(followers FollowersRepository, inbox InboxRepository) Storage {
	return &storage{followers: followers, inbox: inbox}
}

func (s *storage) Followers() FollowersRepository { return s.followers }
func (s *storage) Inbox() InboxRepository         { return s.inbox }
