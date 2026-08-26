package federation

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/solargate/grom/internal/federation/httpsig"
	"github.com/solargate/grom/internal/storage/blob"
)

var (
	ErrUnauthorizedFederation = errors.New("federation request unauthorized")
	ErrActorKeyMismatch       = errors.New("activity actor does not match signature key owner")
)

// AuthenticateRequest verifies HTTP Signatures on an inbound federation request.
func AuthenticateRequest(req *http.Request, body []byte, keys KeyResolver) (actorOwner string, keyID string, err error) {
	if keys == nil {
		return "", "", fmt.Errorf("key resolver required")
	}
	keyID, err = httpsig.KeyID(req)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrUnauthorizedFederation, err)
	}
	pub, owner, err := keys.Resolve(keyID)
	if err != nil {
		if errors.Is(err, errKeyGone) {
			return owner, keyID, err
		}
		return "", keyID, fmt.Errorf("%w: resolve key: %v", ErrUnauthorizedFederation, err)
	}
	if _, verifyErr := httpsig.Verify(req, body, pub); verifyErr != nil {
		if resolver, ok := keys.(interface {
			ResolveFresh(string) (any, string, error)
		}); ok {
			pub, owner, err = resolver.ResolveFresh(keyID)
			if err != nil {
				return "", keyID, fmt.Errorf("%w: %v", ErrUnauthorizedFederation, verifyErr)
			}
			if _, verifyErr = httpsig.Verify(req, body, pub); verifyErr != nil {
				return "", keyID, fmt.Errorf("%w: %v", ErrUnauthorizedFederation, verifyErr)
			}
		} else {
			return "", keyID, fmt.Errorf("%w: %v", ErrUnauthorizedFederation, verifyErr)
		}
	}
	return owner, keyID, nil
}

// AuthenticateActivity verifies the request and ensures activity.actor matches the key owner.
func AuthenticateActivity(req *http.Request, body []byte, activity map[string]any, keys KeyResolver) error {
	actor, _ := activity["actor"].(string)
	owner, keyID, err := AuthenticateRequest(req, body, keys)
	if err != nil {
		if errors.Is(err, errKeyGone) {
			actType, _ := activity["type"].(string)
			if actType == "Delete" && actor != "" && sameHost(actor, keyID) {
				obj := objectIDString(activity)
				if obj == actor || obj == "" {
					return nil
				}
			}
		}
		return err
	}
	if !ownerMatchesKey(actor, owner, keyID) {
		slog.Warn("federation signature actor mismatch",
			"actor", actor,
			"key_owner", owner,
			"key_id", keyID,
		)
		return ErrActorKeyMismatch
	}
	return nil
}

// StaticKeyResolver is a test helper resolver with a fixed key map.
type StaticKeyResolver struct {
	Keys map[string]StaticKey
}

type StaticKey struct {
	Public any
	Owner  string
}

func (s StaticKeyResolver) Resolve(keyID string) (pub any, owner string, err error) {
	k, ok := s.Keys[keyID]
	if !ok {
		return nil, "", fmt.Errorf("unknown keyId %s", keyID)
	}
	return k.Public, k.Owner, nil
}

func ensureInstanceActorKey(blobs blob.Store) error {
	if blobs == nil {
		return nil
	}
	_, err := LoadOrCreateInstanceActorKey(blobs)
	return err
}

func nicknameFromActorURL(actor string) string {
	actor = strings.TrimSuffix(actor, "/")
	const marker = "/users/"
	idx := strings.LastIndex(actor, marker)
	if idx < 0 {
		return ""
	}
	rest := actor[idx+len(marker):]
	rest = strings.SplitN(rest, "/", 2)[0]
	rest = strings.SplitN(rest, "#", 2)[0]
	return rest
}
