package v1

import (
	"strings"

	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/social"
)

func (a *App) localAvatarFieldsForUser(nickname string) (hasAvatar bool, avatarURL string) {
	return avatars.FieldsStore(a.Blobs, nickname)
}

func (a *App) remoteFollowAvatarFields(viewerNickname string, follow *social.Follow) (bool, string) {
	if follow == nil || follow.TargetIsLocal {
		return false, ""
	}

	if a.Federation.Inbox() == nil {
		return follow.TargetAvatarURL != "", follow.TargetAvatarURL
	}

	hasAvatar, avatarURL := a.Federation.Inbox().AuthorAvatarFields(viewerNickname, follow.TargetHandle)
	if hasAvatar {
		return true, avatarURL
	}

	remoteURL := follow.TargetAvatarURL
	if !strings.HasPrefix(remoteURL, "http://") && !strings.HasPrefix(remoteURL, "https://") {
		return false, ""
	}

	_ = a.Federation.Inbox().EnsureAuthor(
		viewerNickname,
		follow.TargetHandle,
		follow.TargetNickname,
		follow.TargetName,
		remoteURL,
		false,
	)
	return a.Federation.Inbox().AuthorAvatarFields(viewerNickname, follow.TargetHandle)
}

func (a *App) cacheRemoteFollowAvatar(viewerNickname string, follow *social.Follow) {
	if follow == nil || follow.TargetIsLocal || follow.TargetAvatarURL == "" {
		return
	}
	if a.Federation.Inbox() == nil {
		return
	}
	_ = a.Federation.Inbox().EnsureAuthor(
		viewerNickname,
		follow.TargetHandle,
		follow.TargetNickname,
		follow.TargetName,
		follow.TargetAvatarURL,
		true,
	)
}

func (a *App) remoteFollowerAvatarFields(viewerNickname string, follower *social.Follower) (bool, string) {
	if follower == nil || follower.FollowerIsLocal {
		return follower.FollowerHasAvatar, follower.FollowerAvatarURL
	}

	if a.Federation.Inbox() == nil {
		return false, ""
	}

	hasAvatar, avatarURL := a.Federation.Inbox().AuthorAvatarFields(viewerNickname, follower.FollowerHandle)
	if hasAvatar {
		return true, avatarURL
	}

	if a.federationDelivery == nil {
		return false, ""
	}

	parsed, err := a.Social.ParseHandle(follower.FollowerHandle)
	if err != nil || parsed.IsLocal {
		return false, ""
	}

	remote, err := a.federationDelivery.ResolveRemote(parsed)
	if err != nil || remote.AvatarURL == "" {
		return false, ""
	}

	name := remote.Name
	if name == "" {
		name = follower.FollowerName
	}

	_ = a.Federation.Inbox().EnsureAuthor(
		viewerNickname,
		follower.FollowerHandle,
		follower.FollowerNickname,
		name,
		remote.AvatarURL,
		false,
	)
	return a.Federation.Inbox().AuthorAvatarFields(viewerNickname, follower.FollowerHandle)
}
