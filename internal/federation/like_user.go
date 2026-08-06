package federation

import (
	"strings"

	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/workouts"
)

func localFederationDomain() string {
	domain := strings.TrimSpace(config.Cfg.Federation.Domain)
	if domain == "" {
		return "localhost"
	}
	return domain
}

// HandleIsLocal reports whether handle belongs to this instance's federation domain.
func HandleIsLocal(handle string) bool {
	domain := domainFromHandle(handle)
	if domain == "" {
		return false
	}
	return strings.EqualFold(domain, localFederationDomain())
}

func publicAvatarURLForHandle(handle, nickname string) string {
	domain := domainFromHandle(handle)
	if domain == "" {
		return ""
	}
	if nickname == "" {
		nickname = OwnerNicknameFromKey(OwnerKeyFromHandle(handle))
	}
	if nickname == "" {
		return ""
	}
	return avatars.PublicURL(domain, nickname)
}

// NormalizeRemoteAvatarURL turns relative avatar paths into absolute public URLs
// for remote handles so other instances can fetch them without guessing the host.
func NormalizeRemoteAvatarURL(handle, nickname, avatarURL string) string {
	avatarURL = strings.TrimSpace(avatarURL)
	if strings.HasPrefix(avatarURL, "http://") || strings.HasPrefix(avatarURL, "https://") {
		return avatarURL
	}
	if HandleIsLocal(handle) {
		return ""
	}
	return publicAvatarURLForHandle(handle, nickname)
}

// ExportLikeUserAvatarURL returns an absolute avatar URL suitable for federation payloads.
func ExportLikeUserAvatarURL(user workouts.WorkoutLikeUser) string {
	u := strings.TrimSpace(user.AvatarURL)
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}
	if u == "" {
		return ""
	}
	return publicAvatarURLForHandle(user.Handle, user.Nickname)
}

func parseFederatedLikeUser(obj map[string]any) workouts.WorkoutLikeUser {
	handle := stringValue(obj, "handle")
	nickname := stringValue(obj, "nickname")
	name := stringValue(obj, "name")
	avatarURL := firstString(obj, "avatarUrl", "avatar_url")
	isLocal := HandleIsLocal(handle)
	if !isLocal {
		avatarURL = NormalizeRemoteAvatarURL(handle, nickname, avatarURL)
	} else {
		// Locality is derived from handle domain on this instance; drop foreign relative paths.
		avatarURL = ""
	}
	return workouts.WorkoutLikeUser{
		Handle:    handle,
		Nickname:  nickname,
		Name:      name,
		IsLocal:   isLocal,
		AvatarURL: avatarURL,
	}
}
