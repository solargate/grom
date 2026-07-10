package v1

import (
	"strings"

	"github.com/solargate/travka/internal/avatars"
	"github.com/solargate/travka/internal/config"
	"github.com/solargate/travka/internal/social"
)

func localAvatarFieldsForUser(nickname string) (hasAvatar bool, avatarURL string) {
	return avatars.Fields(config.Cfg.Data.ResolvedDir, nickname)
}

func publicAvatarURL(nickname string) string {
	domain := config.Cfg.Federation.Domain
	return avatars.PublicURL(domain, nickname)
}

func remoteFollowAvatarFields(viewerNickname string, follow *social.Follow) (bool, string) {
	if follow == nil || follow.TargetIsLocal {
		return false, ""
	}

	if err := initFederation(); err != nil || workoutInboxStore == nil {
		return follow.TargetAvatarURL != "", follow.TargetAvatarURL
	}

	hasAvatar, avatarURL := workoutInboxStore.AuthorAvatarFields(viewerNickname, follow.TargetHandle)
	if hasAvatar {
		return true, avatarURL
	}

	remoteURL := follow.TargetAvatarURL
	if !strings.HasPrefix(remoteURL, "http://") && !strings.HasPrefix(remoteURL, "https://") {
		return false, ""
	}

	_ = workoutInboxStore.EnsureAuthor(
		viewerNickname,
		follow.TargetHandle,
		follow.TargetNickname,
		follow.TargetName,
		remoteURL,
		false,
	)
	return workoutInboxStore.AuthorAvatarFields(viewerNickname, follow.TargetHandle)
}

func cacheRemoteFollowAvatar(viewerNickname string, follow *social.Follow) {
	if follow == nil || follow.TargetIsLocal || follow.TargetAvatarURL == "" {
		return
	}
	if err := initFederation(); err != nil || workoutInboxStore == nil {
		return
	}
	_ = workoutInboxStore.EnsureAuthor(
		viewerNickname,
		follow.TargetHandle,
		follow.TargetNickname,
		follow.TargetName,
		follow.TargetAvatarURL,
		true,
	)
}

func remoteFollowerAvatarFields(viewerNickname string, follower *social.Follower) (bool, string) {
	if follower == nil || follower.FollowerIsLocal {
		return follower.FollowerHasAvatar, follower.FollowerAvatarURL
	}

	if err := initFederation(); err != nil || workoutInboxStore == nil {
		return false, ""
	}

	hasAvatar, avatarURL := workoutInboxStore.AuthorAvatarFields(viewerNickname, follower.FollowerHandle)
	if hasAvatar {
		return true, avatarURL
	}

	if err := initSocialService(); err != nil || federationDelivery == nil {
		return false, ""
	}

	parsed, err := socialService.ParseHandle(follower.FollowerHandle)
	if err != nil || parsed.IsLocal {
		return false, ""
	}

	remote, err := federationDelivery.ResolveRemote(parsed)
	if err != nil || remote.AvatarURL == "" {
		return false, ""
	}

	name := remote.Name
	if name == "" {
		name = follower.FollowerName
	}

	_ = workoutInboxStore.EnsureAuthor(
		viewerNickname,
		follower.FollowerHandle,
		follower.FollowerNickname,
		name,
		remote.AvatarURL,
		false,
	)
	return workoutInboxStore.AuthorAvatarFields(viewerNickname, follower.FollowerHandle)
}
