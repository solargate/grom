package federation

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
)

func hasFederatedAvatar(ctx context.Context, blobs blob.Store, viewerNickname, ownerKey string) bool {
	if blobs == nil {
		return false
	}
	exists, err := blobs.Exists(ctx, keys.FederatedInboxAvatar(viewerNickname, ownerKey))
	return err == nil && exists
}

func FederatedAvatarAPIPath(handle string, version int) string {
	path := fmt.Sprintf("/api/v1/federation/authors/%s/avatar", OwnerKeyFromHandle(handle))
	if version > 0 {
		return fmt.Sprintf("%s?v=%d", path, version)
	}
	return path
}

func effectiveRemoteAvatarURL(meta AuthorMeta) string {
	if meta.RemoteAvatarURL != "" {
		return meta.RemoteAvatarURL
	}
	if strings.HasPrefix(meta.AvatarURL, "http://") || strings.HasPrefix(meta.AvatarURL, "https://") {
		return meta.AvatarURL
	}
	return ""
}

func authorAvatarFields(blobs blob.Store, viewerNickname, ownerKey, handle string, meta AuthorMeta) (hasAvatar bool, avatarURL string) {
	ctx := context.Background()
	if hasFederatedAvatar(ctx, blobs, viewerNickname, ownerKey) {
		return true, FederatedAvatarAPIPath(handle, meta.AvatarVersion)
	}
	if effectiveRemoteAvatarURL(meta) != "" {
		return true, FederatedAvatarAPIPath(handle, meta.AvatarVersion)
	}
	return false, ""
}

func cacheRemoteAvatar(client *http.Client, blobs blob.Store, viewerNickname, ownerKey, remoteURL string) error {
	if client == nil || remoteURL == "" || blobs == nil {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, remoteURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("avatar fetch failed: %s", resp.Status)
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, avatars.MaxUploadBytes+1))
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return avatars.ErrInvalidAvatar
	}
	if len(raw) > avatars.MaxUploadBytes {
		return avatars.ErrAvatarTooLarge
	}

	avatarKey := keys.FederatedInboxAvatar(viewerNickname, ownerKey)
	return avatars.SaveKey(blobs, avatarKey, raw)
}

func syncAuthorAvatar(client *http.Client, blobs blob.Store, viewerNickname, ownerKey, handle string, meta *AuthorMeta, remoteURL string, refresh bool) {
	ctx := context.Background()
	prevRemote := effectiveRemoteAvatarURL(*meta)
	if remoteURL != "" {
		meta.RemoteAvatarURL = remoteURL
	}

	remote := effectiveRemoteAvatarURL(*meta)
	if remote == "" {
		if hasFederatedAvatar(ctx, blobs, viewerNickname, ownerKey) {
			meta.AvatarURL = FederatedAvatarAPIPath(handle, meta.AvatarVersion)
		} else {
			meta.AvatarURL = ""
		}
		return
	}

	needsFetch := !hasFederatedAvatar(ctx, blobs, viewerNickname, ownerKey) || (remoteURL != "" && remoteURL != prevRemote) || refresh
	if client != nil && needsFetch {
		if err := cacheRemoteAvatar(client, blobs, viewerNickname, ownerKey, remote); err != nil {
			slog.Warn("federated avatar cache failed", "handle", handle, "err", err)
		} else {
			meta.AvatarVersion++
		}
	}

	if hasFederatedAvatar(ctx, blobs, viewerNickname, ownerKey) {
		meta.AvatarURL = FederatedAvatarAPIPath(handle, meta.AvatarVersion)
	} else {
		meta.AvatarURL = ""
	}
}
