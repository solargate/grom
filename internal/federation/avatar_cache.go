package federation

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/data"
)

func federatedAvatarPath(ownerDir string) string {
	return filepath.Join(ownerDir, data.AvatarFileName)
}

func hasFederatedAvatar(ownerDir string) bool {
	_, err := os.Stat(federatedAvatarPath(ownerDir))
	return err == nil
}

func FederatedAvatarAPIPath(handle string, version int) string {
	path := fmt.Sprintf("/api/v1/federation/authors/%s/avatar", ownerDirName(handle))
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

func authorAvatarFields(ownerDir, handle string, meta AuthorMeta) (hasAvatar bool, avatarURL string) {
	if hasFederatedAvatar(ownerDir) {
		return true, FederatedAvatarAPIPath(handle, meta.AvatarVersion)
	}
	if effectiveRemoteAvatarURL(meta) != "" {
		return true, FederatedAvatarAPIPath(handle, meta.AvatarVersion)
	}
	return false, ""
}

func cacheRemoteAvatar(client *http.Client, ownerDir, remoteURL string) error {
	if client == nil || remoteURL == "" {
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

	return avatars.SaveFile(federatedAvatarPath(ownerDir), raw)
}

func syncAuthorAvatar(client *http.Client, ownerDir, handle string, meta *AuthorMeta, remoteURL string, refresh bool) {
	prevRemote := effectiveRemoteAvatarURL(*meta)
	if remoteURL != "" {
		meta.RemoteAvatarURL = remoteURL
	}

	remote := effectiveRemoteAvatarURL(*meta)
	if remote == "" {
		if hasFederatedAvatar(ownerDir) {
			meta.AvatarURL = FederatedAvatarAPIPath(handle, meta.AvatarVersion)
		} else {
			meta.AvatarURL = ""
		}
		return
	}

	needsFetch := !hasFederatedAvatar(ownerDir) || (remoteURL != "" && remoteURL != prevRemote) || refresh
	if client != nil && needsFetch {
		if err := cacheRemoteAvatar(client, ownerDir, remote); err != nil {
			log.Printf("federated avatar cache failed for %s: %v", handle, err)
		} else {
			meta.AvatarVersion++
		}
	}

	if hasFederatedAvatar(ownerDir) {
		meta.AvatarURL = FederatedAvatarAPIPath(handle, meta.AvatarVersion)
	} else {
		meta.AvatarURL = ""
	}
}
