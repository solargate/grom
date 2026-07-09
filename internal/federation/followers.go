package federation

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/solargate/travka/internal/data"
	"gopkg.in/yaml.v3"
)

type InboundFollower struct {
	ActorURI string `yaml:"actor_uri" json:"actor_uri"`
	Inbox    string `yaml:"inbox" json:"inbox"`
	Handle   string `yaml:"handle" json:"handle"`
}

type followersFile struct {
	Followers []InboundFollower `yaml:"followers"`
}

type FollowersStore struct {
	dataDir string
	mu      sync.Mutex
}

func NewFollowersStore(dataDir string) *FollowersStore {
	return &FollowersStore{dataDir: dataDir}
}

func (s *FollowersStore) path(nickname string) string {
	return filepath.Join(data.UserDir(s.dataDir, nickname), "federation", "followers.yaml")
}

func (s *FollowersStore) Add(nickname string, follower InboundFollower) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.path(nickname)
	var file followersFile
	data, err := os.ReadFile(path)
	if err == nil {
		_ = yaml.Unmarshal(data, &file)
	} else if !os.IsNotExist(err) {
		return err
	}

	for i := range file.Followers {
		if file.Followers[i].ActorURI == follower.ActorURI {
			return nil
		}
	}
	file.Followers = append(file.Followers, follower)
	return s.save(path, file)
}

func (s *FollowersStore) ListInboxes(nickname string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.path(nickname)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var file followersFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse followers: %w", err)
	}
	inboxes := make([]string, 0, len(file.Followers))
	for i := range file.Followers {
		if file.Followers[i].Inbox != "" {
			inboxes = append(inboxes, file.Followers[i].Inbox)
		}
	}
	return inboxes, nil
}

func (s *FollowersStore) save(path string, file followersFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(&file)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
