package social

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solargate/travka/internal/config"
	"github.com/solargate/travka/internal/users"
)

var (
	ErrInvalidHandle   = errors.New("invalid user handle")
	ErrUserNotFound    = errors.New("user not found")
	ErrRemoteNotReady  = errors.New("remote federation is not enabled on this server")
)

type Delivery interface {
	DeliverFollow(follow *Follow) error
	DeliverUndo(follow *Follow) error
}

type noopDelivery struct{}

func (noopDelivery) DeliverFollow(*Follow) error { return nil }
func (noopDelivery) DeliverUndo(*Follow) error   { return nil }

type Service struct {
	users    *users.Store
	follows  *Store
	domain   string
	enabled  bool
	delivery Delivery
}

func NewService(userStore *users.Store, followStore *Store) *Service {
	domain := config.Cfg.Federation.Domain
	if domain == "" {
		domain = "localhost"
	}
	return &Service{
		users:    userStore,
		follows:  followStore,
		domain:   domain,
		enabled:  config.Cfg.Federation.Enabled,
		delivery: noopDelivery{},
	}
}

func (s *Service) SetDelivery(d Delivery) {
	if d != nil {
		s.delivery = d
	}
}

func (s *Service) LocalDomain() string {
	return s.domain
}

func (s *Service) LocalHandle(nickname string) string {
	return nickname + "@" + s.domain
}

func (s *Service) ActorURI(nickname string) string {
	return fmt.Sprintf("https://%s/users/%s", s.domain, nickname)
}

type ParsedHandle struct {
	Nickname string
	Domain   string
	Handle   string
	IsLocal  bool
}

func (s *Service) ParseHandle(raw string) (ParsedHandle, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "@")
	if raw == "" {
		return ParsedHandle{}, ErrInvalidHandle
	}

	at := strings.LastIndex(raw, "@")
	if at < 0 {
		nickname := raw
		if err := validateNicknamePart(nickname); err != nil {
			return ParsedHandle{}, err
		}
		handle := s.LocalHandle(nickname)
		return ParsedHandle{
			Nickname: nickname,
			Domain:   s.domain,
			Handle:   handle,
			IsLocal:  true,
		}, nil
	}

	nickname := raw[:at]
	domain := raw[at+1:]
	if nickname == "" || domain == "" {
		return ParsedHandle{}, ErrInvalidHandle
	}
	if err := validateNicknamePart(nickname); err != nil {
		return ParsedHandle{}, err
	}

	handle := nickname + "@" + domain
	return ParsedHandle{
		Nickname: nickname,
		Domain:   domain,
		Handle:   handle,
		IsLocal:  strings.EqualFold(domain, s.domain),
	}, nil
}

func validateNicknamePart(nickname string) error {
	if nickname == "" || strings.Contains(nickname, "/") || strings.Contains(nickname, "\\") {
		return ErrInvalidHandle
	}
	return nil
}

type UserSearchResult struct {
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Handle   string `json:"handle"`
	IsLocal  bool   `json:"is_local"`
}

func (s *Service) SearchLocal(query string, excludeUserID string) ([]UserSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	if strings.Contains(query, "@") {
		parsed, err := s.ParseHandle(query)
		if err != nil {
			return nil, err
		}
		if !parsed.IsLocal {
			if !s.enabled {
				return nil, ErrRemoteNotReady
			}
			return s.resolveRemotePreview(parsed)
		}
		user, err := s.users.FindByNickname(parsed.Nickname)
		if err != nil {
			return nil, ErrUserNotFound
		}
		if user.ID == excludeUserID {
			return nil, nil
		}
		return []UserSearchResult{toSearchResult(user, s.LocalHandle(user.Nickname), true)}, nil
	}

	usersList, err := s.users.Search(query, excludeUserID, 20)
	if err != nil {
		return nil, err
	}
	result := make([]UserSearchResult, 0, len(usersList))
	for i := range usersList {
		result = append(result, toSearchResult(&usersList[i], s.LocalHandle(usersList[i].Nickname), true))
	}
	return result, nil
}

func toSearchResult(user *users.User, handle string, isLocal bool) UserSearchResult {
	return UserSearchResult{
		Nickname: user.Nickname,
		Name:     user.Name,
		Handle:   handle,
		IsLocal:  isLocal,
	}
}

func (s *Service) resolveRemotePreview(parsed ParsedHandle) ([]UserSearchResult, error) {
	if s.delivery == nil {
		return nil, ErrRemoteNotReady
	}
	type remoteResolver interface {
		ResolveRemote(parsed ParsedHandle) (*UserSearchResult, error)
	}
	if r, ok := s.delivery.(remoteResolver); ok {
		res, err := r.ResolveRemote(parsed)
		if err != nil {
			return nil, err
		}
		return []UserSearchResult{*res}, nil
	}
	return nil, ErrRemoteNotReady
}

type FollowResponse struct {
	Follow
}

func (s *Service) Follow(followerID, rawHandle string) (*Follow, error) {
	parsed, err := s.ParseHandle(rawHandle)
	if err != nil {
		return nil, err
	}

	follower, err := s.users.FindByID(followerID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(parsed.Nickname, follower.Nickname) && parsed.IsLocal {
		return nil, ErrCannotFollowSelf
	}

	if existing, err := s.follows.FindExisting(followerID, parsed.Handle); err == nil {
		return existing, nil
	}

	if parsed.IsLocal {
		target, err := s.users.FindByNickname(parsed.Nickname)
		if err != nil {
			return nil, ErrUserNotFound
		}
		follow := Follow{
			FollowerID:     followerID,
			TargetActorURI: s.ActorURI(target.Nickname),
			TargetHandle:   parsed.Handle,
			TargetNickname: target.Nickname,
			TargetName:     target.Name,
			TargetIsLocal:  true,
			Status:         StatusActive,
			CreatedAt:      time.Now().UTC(),
		}
		return s.follows.Create(follow)
	}

	if !s.enabled {
		return nil, ErrRemoteNotReady
	}

	follow := Follow{
		FollowerID:     followerID,
		TargetActorURI: fmt.Sprintf("https://%s/users/%s", parsed.Domain, parsed.Nickname),
		TargetHandle:   parsed.Handle,
		TargetNickname: parsed.Nickname,
		Status:         StatusPending,
		CreatedAt:      time.Now().UTC(),
	}
	created, err := s.follows.Create(follow)
	if err != nil {
		return nil, err
	}
	activityID := fmt.Sprintf("%s/follows/%s", s.ActorURI(follower.Nickname), uuid.NewString())
	created, err = s.follows.UpdateActivityID(created.ID, activityID)
	if err != nil {
		_ = s.follows.Delete(created.ID)
		return nil, err
	}
	if err := s.delivery.DeliverFollow(created); err != nil {
		_ = s.follows.Delete(created.ID)
		return nil, err
	}
	return created, nil
}

func (s *Service) Unfollow(followerID, followID string) error {
	follow, err := s.follows.FindByID(followID)
	if err != nil {
		return err
	}
	if follow.FollowerID != followerID {
		return ErrFollowNotFound
	}
	if !follow.TargetIsLocal {
		if err := s.delivery.DeliverUndo(follow); err != nil {
			return err
		}
	}
	return s.follows.Delete(followID)
}

func (s *Service) ListFollowing(followerID string) ([]Follow, error) {
	return s.follows.ListByFollower(followerID)
}

func (s *Service) ActiveFollowingNicknames(followerID string) ([]string, error) {
	follows, err := s.follows.ListActiveFollowing(followerID)
	if err != nil {
		return nil, err
	}
	nicknames := make([]string, 0, len(follows))
	for i := range follows {
		nicknames = append(nicknames, follows[i].TargetNickname)
	}
	return nicknames, nil
}

func (s *Service) ActivateFollowByActivityID(activityID string) error {
	follow, err := s.follows.FindByFollowActivityID(activityID)
	if err != nil {
		return err
	}
	_, err = s.follows.UpdateStatus(follow.ID, StatusActive)
	return err
}

func (s *Service) HandleIncomingAccept(followActivityID string) error {
	// Updated when federation inbox is wired.
	return nil
}
