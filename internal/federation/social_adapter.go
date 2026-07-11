package federation

import (
	"strings"

	"github.com/solargate/grom/internal/social"
)

type inboundFollowersAdapter struct {
	store *FollowersStore
}

func NewInboundFollowersAdapter(store *FollowersStore) social.InboundFollowersSource {
	return inboundFollowersAdapter{store: store}
}

func (a inboundFollowersAdapter) ListInboundFollowers(nickname string) ([]social.InboundFollowerInfo, error) {
	followers, err := a.store.List(nickname)
	if err != nil {
		return nil, err
	}
	result := make([]social.InboundFollowerInfo, 0, len(followers))
	for i := range followers {
		handle := followers[i].Handle
		if strings.HasPrefix(handle, "https://") {
			handle = actorToHandle(handle)
		}
		if handle == "" {
			continue
		}
		nick := handle
		if at := strings.Index(handle, "@"); at >= 0 {
			nick = handle[:at]
		}
		result = append(result, social.InboundFollowerInfo{
			Handle:   handle,
			Nickname: nick,
		})
	}
	return result, nil
}
