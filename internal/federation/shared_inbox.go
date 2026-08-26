package federation

import (
	"strings"
)

// ResolveSharedInboxRecipients finds local nicknames that should process an activity
// delivered to the shared inbox.
func (p *InboxProcessor) ResolveSharedInboxRecipients(activity map[string]any) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(nick string) {
		nick = strings.TrimSpace(nick)
		if nick == "" {
			return
		}
		if _, ok := seen[nick]; ok {
			return
		}
		seen[nick] = struct{}{}
		out = append(out, nick)
	}

	for _, uri := range collectAddressedURIs(activity) {
		if nick, ok := LocalActorNickname(uri); ok {
			add(nick)
		}
	}

	actType, _ := activity["type"].(string)
	switch actType {
	case "Follow":
		if nick, ok := LocalActorNickname(objectIDString(activity)); ok {
			add(nick)
		}
		if obj, ok := activity["object"].(map[string]any); ok {
			if id, _ := obj["id"].(string); id != "" {
				if nick, ok := LocalActorNickname(id); ok {
					add(nick)
				}
			}
			if id, _ := obj["object"].(string); id != "" {
				if nick, ok := LocalActorNickname(id); ok {
					add(nick)
				}
			}
		}
	case "Accept":
		// Accept activates a pending Follow owned by the original follower (object.actor).
		// Do not rely on ListActiveByTarget — the follow is still pending.
		if obj, ok := activity["object"].(map[string]any); ok {
			if actor, _ := obj["actor"].(string); actor != "" {
				if nick, ok := LocalActorNickname(actor); ok {
					add(nick)
				}
			}
			if followID, _ := obj["id"].(string); followID != "" && p.social != nil {
				if nick, err := p.social.LocalNicknameForFollowActivityID(followID); err == nil {
					add(nick)
				}
			}
		} else if followID := objectIDString(activity); followID != "" && p.social != nil {
			if nick, err := p.social.LocalNicknameForFollowActivityID(followID); err == nil {
				add(nick)
			}
		}
	case "Undo":
		if obj, ok := activity["object"].(map[string]any); ok {
			if obj["type"] == "Follow" {
				if id, _ := obj["object"].(string); id != "" {
					if nick, ok := LocalActorNickname(id); ok {
						add(nick)
					}
				}
			}
		}
	}

	// Public / followers deliveries: local users following the remote actor.
	if p.social != nil {
		actor, _ := activity["actor"].(string)
		handle := actorToHandle(actor)
		if handle != "" {
			if nicks, err := p.social.ListLocalNicknamesFollowingHandle(handle); err == nil {
				for _, nick := range nicks {
					add(nick)
				}
			}
		}
	}

	return out
}
