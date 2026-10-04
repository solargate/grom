package file

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/solargate/grom/internal/notifications"
	"gopkg.in/yaml.v3"
)

const pushSubscriptionsFileName = "push_subscriptions.yaml"

type pushSubscriptionsFile struct {
	Subscriptions []notifications.PushSubscription `yaml:"subscriptions"`
}

// PushSubscriptionStore persists Web Push subscriptions as YAML.
type PushSubscriptionStore struct {
	path string
	mu   sync.Mutex
}

func NewPushSubscriptionStore(dataDir string) *PushSubscriptionStore {
	return &PushSubscriptionStore{path: filepath.Join(dataDir, pushSubscriptionsFileName)}
}

func (s *PushSubscriptionStore) Upsert(sub notifications.PushSubscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.load()
	if err != nil {
		return err
	}
	found := false
	for i := range list {
		if list[i].UserID == sub.UserID && list[i].InstallationID == sub.InstallationID {
			list[i] = sub
			found = true
			break
		}
	}
	if !found {
		list = append(list, sub)
	}
	return s.save(list)
}

func (s *PushSubscriptionStore) ListByUser(userID string) ([]notifications.PushSubscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]notifications.PushSubscription, 0)
	for _, sub := range list {
		if sub.UserID == userID {
			out = append(out, sub)
		}
	}
	return out, nil
}

func (s *PushSubscriptionStore) DeleteByUserAndInstallation(userID, installationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.load()
	if err != nil {
		return err
	}
	out := list[:0]
	found := false
	for _, sub := range list {
		if sub.UserID == userID && sub.InstallationID == installationID {
			found = true
			continue
		}
		out = append(out, sub)
	}
	if !found {
		return notifications.ErrNotFound
	}
	return s.save(out)
}

func (s *PushSubscriptionStore) DeleteAllForUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.load()
	if err != nil {
		return err
	}
	out := list[:0]
	for _, sub := range list {
		if sub.UserID == userID {
			continue
		}
		out = append(out, sub)
	}
	return s.save(out)
}

func (s *PushSubscriptionStore) DeleteByEndpoint(endpoint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.load()
	if err != nil {
		return err
	}
	out := list[:0]
	for _, sub := range list {
		if sub.Endpoint == endpoint {
			continue
		}
		out = append(out, sub)
	}
	return s.save(out)
}

// ListAll returns every push subscription (migration).
func (s *PushSubscriptionStore) ListAll() ([]notifications.PushSubscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]notifications.PushSubscription, len(list))
	copy(out, list)
	return out, nil
}

// Import writes a subscription as-is (storage migration).
func (s *PushSubscriptionStore) Import(sub notifications.PushSubscription) error {
	if sub.UpdatedAt.IsZero() {
		sub.UpdatedAt = time.Now().UTC()
	}
	return s.Upsert(sub)
}

func (s *PushSubscriptionStore) load() ([]notifications.PushSubscription, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var file pushSubscriptionsFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return file.Subscriptions, nil
}

func (s *PushSubscriptionStore) save(list []notifications.PushSubscription) error {
	if list == nil {
		list = []notifications.PushSubscription{}
	}
	data, err := yaml.Marshal(pushSubscriptionsFile{Subscriptions: list})
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

var _ notifications.Repository = (*PushSubscriptionStore)(nil)
