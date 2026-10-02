package notifications

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// VAPIDKeys holds a Web Push VAPID key pair.
type VAPIDKeys struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

const vapidFileName = "vapid.json"

// LoadOrCreateVAPID loads VAPID keys from dataDir/notifications/vapid.json,
// generating and persisting a new pair when the file is missing.
func LoadOrCreateVAPID(dataDir string) (*VAPIDKeys, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("data dir is required for VAPID keys")
	}
	dir := filepath.Join(dataDir, "notifications")
	path := filepath.Join(dir, vapidFileName)

	data, err := os.ReadFile(path)
	if err == nil {
		var keys VAPIDKeys
		if err := json.Unmarshal(data, &keys); err != nil {
			return nil, fmt.Errorf("parse vapid keys: %w", err)
		}
		if keys.PublicKey == "" || keys.PrivateKey == "" {
			return nil, fmt.Errorf("vapid keys file is incomplete")
		}
		return &keys, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read vapid keys: %w", err)
	}

	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		return nil, fmt.Errorf("generate vapid keys: %w", err)
	}
	keys := &VAPIDKeys{PublicKey: publicKey, PrivateKey: privateKey}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create notifications dir: %w", err)
	}
	payload, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return nil, err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0600); err != nil {
		return nil, fmt.Errorf("write vapid keys: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return nil, fmt.Errorf("persist vapid keys: %w", err)
	}
	return keys, nil
}

// VAPIDStore holds process-wide VAPID keys loaded at startup.
type VAPIDStore struct {
	mu   sync.RWMutex
	keys *VAPIDKeys
}

func NewVAPIDStore(keys *VAPIDKeys) *VAPIDStore {
	return &VAPIDStore{keys: keys}
}

func (s *VAPIDStore) PublicKey() string {
	if s == nil {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.keys == nil {
		return ""
	}
	return s.keys.PublicKey
}

func (s *VAPIDStore) Keys() *VAPIDKeys {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.keys
}
