package federation

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/solargate/grom/internal/storage/blob"
	"gopkg.in/yaml.v3"
)

const authorMetaFileName = "author.yaml"

type AuthorMeta struct {
	Nickname        string `yaml:"nickname" json:"nickname"`
	Handle          string `yaml:"handle" json:"handle"`
	Name            string `yaml:"name,omitempty" json:"name,omitempty"`
	AvatarURL       string `yaml:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	RemoteAvatarURL string `yaml:"remote_avatar_url,omitempty" json:"remote_avatar_url,omitempty"`
	AvatarVersion   int    `yaml:"avatar_version,omitempty" json:"avatar_version,omitempty"`
}

func readAuthorMeta(ownerDir string) (AuthorMeta, error) {
	path := filepath.Join(ownerDir, authorMetaFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return AuthorMeta{}, nil
		}
		return AuthorMeta{}, err
	}
	var meta AuthorMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return AuthorMeta{}, err
	}
	return meta, nil
}

func writeAuthorMeta(ownerDir string, meta AuthorMeta) error {
	path := filepath.Join(ownerDir, authorMetaFileName)
	data, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func mergeAuthorMeta(ownerDir, handle, nickname string, actor map[string]any, client *http.Client, blobs blob.Store, viewerNickname, ownerKey string) error {
	refresh := false
	if actor != nil && ExtractIconURL(actor) != "" {
		refresh = true
	}
	return mergeAuthorMetaWithRefresh(ownerDir, handle, nickname, actor, client, refresh, blobs, viewerNickname, ownerKey)
}

func mergeAuthorMetaWithRefresh(ownerDir, handle, nickname string, actor map[string]any, client *http.Client, refresh bool, blobs blob.Store, viewerNickname, ownerKey string) error {
	meta, err := readAuthorMeta(ownerDir)
	if err != nil {
		return err
	}
	UpdateAuthorMeta(&meta, handle, nickname, actor, client, refresh, blobs, viewerNickname, ownerKey)
	if meta.Handle == "" && meta.Nickname == "" {
		return nil
	}
	return writeAuthorMeta(ownerDir, meta)
}

// UpdateAuthorMeta merges actor fields and syncs avatar cache into meta in memory.
func UpdateAuthorMeta(meta *AuthorMeta, handle, nickname string, actor map[string]any, client *http.Client, refresh bool, blobs blob.Store, viewerNickname, ownerKey string) {
	if meta == nil {
		return
	}
	if meta.Handle == "" {
		meta.Handle = handle
	}
	if meta.Nickname == "" {
		meta.Nickname = nickname
	}

	remoteURL := ""
	if actor != nil {
		if name := ExtractActorName(actor); name != "" {
			meta.Name = name
		}
		if avatarURL := ExtractIconURL(actor); avatarURL != "" {
			remoteURL = avatarURL
		}
	}

	syncAuthorAvatar(client, blobs, viewerNickname, ownerKey, meta.Handle, meta, remoteURL, refresh)
}

// AuthorAvatarFields returns avatar presence and API URL for a federated author.
func AuthorAvatarFields(blobs blob.Store, viewerNickname, ownerKey, handle string, meta AuthorMeta) (bool, string) {
	return authorAvatarFields(blobs, viewerNickname, ownerKey, handle, meta)
}

func authorActor(name, remoteAvatarURL string) map[string]any {
	if name == "" && remoteAvatarURL == "" {
		return nil
	}
	actor := map[string]any{}
	if name != "" {
		actor["name"] = name
	}
	if remoteAvatarURL != "" {
		actor["icon"] = map[string]any{"url": remoteAvatarURL}
	}
	return actor
}
