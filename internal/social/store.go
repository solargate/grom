package social

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

var (
	ErrFollowNotFound      = errors.New("follow not found")
	ErrAlreadyFollowing    = errors.New("already following this user")
	ErrCannotFollowSelf    = errors.New("cannot follow yourself")
	ErrFollowNotActive     = errors.New("follow is not active")
)

const followsFileName = "follows.yaml"

type Store struct {
	dataDir string
	path    string
	mu      sync.Mutex
	follows []Follow
}

func NewStore(dataDir string) (*Store, error) {
	fedDir := filepath.Join(dataDir, "federation")
	if err := os.MkdirAll(fedDir, 0700); err != nil {
		return nil, fmt.Errorf("create federation dir: %w", err)
	}

	s := &Store{
		dataDir: dataDir,
		path:    filepath.Join(fedDir, followsFileName),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.follows = nil
			return nil
		}
		return err
	}

	var file followsFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("parse follows file: %w", err)
	}
	s.follows = file.Follows
	return nil
}

func (s *Store) save() error {
	file := followsFile{Follows: s.follows}
	data, err := yaml.Marshal(&file)
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) FindByID(id string) (*Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		if s.follows[i].ID == id {
			f := s.follows[i]
			return &f, nil
		}
	}
	return nil, ErrFollowNotFound
}

func (s *Store) ListByFollower(followerID string) ([]Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]Follow, 0)
	for i := range s.follows {
		if s.follows[i].FollowerID == followerID {
			result = append(result, s.follows[i])
		}
	}
	return result, nil
}

func (s *Store) ListActiveFollowing(followerID string) ([]Follow, error) {
	all, err := s.ListByFollower(followerID)
	if err != nil {
		return nil, err
	}
	result := make([]Follow, 0, len(all))
	for i := range all {
		if all[i].Status == StatusActive {
			result = append(result, all[i])
		}
	}
	return result, nil
}

func (s *Store) FindExisting(followerID, targetHandle string) (*Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		f := s.follows[i]
		if f.FollowerID == followerID && f.TargetHandle == targetHandle && f.Status != StatusRejected {
			return &f, nil
		}
	}
	return nil, ErrFollowNotFound
}

func (s *Store) Create(follow Follow) (*Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		f := s.follows[i]
		if f.FollowerID == follow.FollowerID && f.TargetHandle == follow.TargetHandle && f.Status != StatusRejected {
			return nil, ErrAlreadyFollowing
		}
	}

	if follow.ID == "" {
		follow.ID = uuid.NewString()
	}
	s.follows = append(s.follows, follow)
	if err := s.save(); err != nil {
		s.follows = s.follows[:len(s.follows)-1]
		return nil, err
	}
	result := follow
	return &result, nil
}

func (s *Store) UpdateStatus(id, status string) (*Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		if s.follows[i].ID == id {
			s.follows[i].Status = status
			if err := s.save(); err != nil {
				return nil, err
			}
			f := s.follows[i]
			return &f, nil
		}
	}
	return nil, ErrFollowNotFound
}

func (s *Store) FindByFollowActivityID(activityID string) (*Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		if s.follows[i].FollowActivityID == activityID {
			f := s.follows[i]
			return &f, nil
		}
	}
	return nil, ErrFollowNotFound
}

func (s *Store) UpdateActivityID(id, activityID string) (*Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		if s.follows[i].ID == id {
			s.follows[i].FollowActivityID = activityID
			if err := s.save(); err != nil {
				return nil, err
			}
			f := s.follows[i]
			return &f, nil
		}
	}
	return nil, ErrFollowNotFound
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		if s.follows[i].ID == id {
			s.follows = append(s.follows[:i], s.follows[i+1:]...)
			return s.save()
		}
	}
	return ErrFollowNotFound
}
