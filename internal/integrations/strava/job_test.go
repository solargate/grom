package strava

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestJobCleanupAndRestart(t *testing.T) {
	dir := t.TempDir()
	manager := NewJobManager(dir, nil, nil, nil, nil, nil, nil)
	job, err := manager.BeginUpload("user", "alice")
	if err != nil {
		t.Fatal(err)
	}
	sibling := filepath.Join(dir, "alice", "other")
	if err := os.WriteFile(sibling, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.BeginUpload("user", "alice"); err == nil {
		t.Fatal("accepted another upload while the first job was active")
	}

	var polling sync.WaitGroup
	polling.Go(func() {
		for range 50 {
			manager.Status("user")
		}
	})
	manager.CancelUpload(job)
	polling.Wait()
	if _, err := os.Stat(job.WorkDir); !os.IsNotExist(err) {
		t.Fatalf("job directory still exists: %v", err)
	}
	if _, err := os.Stat(sibling); err != nil {
		t.Fatalf("sibling was removed: %v", err)
	}
	if _, err := manager.BeginUpload("user", "alice"); err != nil {
		t.Fatal(err)
	}

	restarted := NewJobManager(dir, nil, nil, nil, nil, nil, nil)
	if status := restarted.Status("user"); status.Active || status.Phase != PhaseFailed {
		t.Fatalf("unexpected recovered status: %+v", status)
	}
	if _, err := restarted.BeginUpload("user", "alice"); err != nil {
		t.Fatal(err)
	}
}
