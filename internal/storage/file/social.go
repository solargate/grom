package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/solargate/grom/internal/social"
	"gopkg.in/yaml.v3"
)

const followsFileName = "follows.yaml"

type followsFile struct {
	Follows []social.Follow `yaml:"follows"`
}

type SocialStore struct {
	dataDir string
	path    string
	mu      sync.Mutex
	follows []social.Follow
}

func NewSocialStore(dataDir string) (*SocialStore, error) {
	fedDir := filepath.Join(dataDir, "federation")
	if err := os.MkdirAll(fedDir, 0700); err != nil {
		return nil, fmt.Errorf("create federation dir: %w", err)
	}

	s := &SocialStore{
		dataDir: dataDir,
		path:    filepath.Join(fedDir, followsFileName),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SocialStore) load() error {
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

func (s *SocialStore) save() error {
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

func (s *SocialStore) FindByID(id string) (*social.Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		if s.follows[i].ID == id {
			f := s.follows[i]
			return &f, nil
		}
	}
	return nil, social.ErrFollowNotFound
}

func (s *SocialStore) ListByFollower(followerID string) ([]social.Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]social.Follow, 0)
	for i := range s.follows {
		if s.follows[i].FollowerID == followerID {
			result = append(result, s.follows[i])
		}
	}
	return result, nil
}

func (s *SocialStore) ListActiveFollowing(followerID string) ([]social.Follow, error) {
	all, err := s.ListByFollower(followerID)
	if err != nil {
		return nil, err
	}
	result := make([]social.Follow, 0, len(all))
	for i := range all {
		if all[i].Status == social.StatusActive {
			result = append(result, all[i])
		}
	}
	return result, nil
}

func (s *SocialStore) ListActiveByTarget(targetHandle string) ([]social.Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]social.Follow, 0)
	for i := range s.follows {
		if s.follows[i].TargetHandle == targetHandle && s.follows[i].Status == social.StatusActive {
			result = append(result, s.follows[i])
		}
	}
	return result, nil
}

func (s *SocialStore) FindExisting(followerID, targetHandle string) (*social.Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		f := s.follows[i]
		if f.FollowerID == followerID && f.TargetHandle == targetHandle && f.Status != social.StatusRejected {
			return &f, nil
		}
	}
	return nil, social.ErrFollowNotFound
}

func (s *SocialStore) Create(follow social.Follow) (*social.Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		f := s.follows[i]
		if f.FollowerID == follow.FollowerID && f.TargetHandle == follow.TargetHandle && f.Status != social.StatusRejected {
			return nil, social.ErrAlreadyFollowing
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

func (s *SocialStore) UpdateStatus(id, status string) (*social.Follow, error) {
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
	return nil, social.ErrFollowNotFound
}

func (s *SocialStore) FindByFollowActivityID(activityID string) (*social.Follow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		if s.follows[i].FollowActivityID == activityID {
			f := s.follows[i]
			return &f, nil
		}
	}
	return nil, social.ErrFollowNotFound
}

func (s *SocialStore) UpdateActivityID(id, activityID string) (*social.Follow, error) {
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
	return nil, social.ErrFollowNotFound
}

func (s *SocialStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.follows {
		if s.follows[i].ID == id {
			s.follows = append(s.follows[:i], s.follows[i+1:]...)
			return s.save()
		}
	}
	return social.ErrFollowNotFound
}

var _ social.Repository = (*SocialStore)(nil)
