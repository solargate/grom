package notifications_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solargate/grom/internal/notifications"
)

func TestLoadOrCreateVAPIDPersists(t *testing.T) {
	dir := t.TempDir()
	keys1, err := notifications.LoadOrCreateVAPID(dir)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if keys1.PublicKey == "" || keys1.PrivateKey == "" {
		t.Fatalf("empty keys: %#v", keys1)
	}
	path := filepath.Join(dir, "notifications", "vapid.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("vapid file missing: %v", err)
	}
	keys2, err := notifications.LoadOrCreateVAPID(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if keys1.PublicKey != keys2.PublicKey || keys1.PrivateKey != keys2.PrivateKey {
		t.Fatalf("keys changed across reload")
	}
}
