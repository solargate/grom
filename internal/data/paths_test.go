package data

import (
	"path/filepath"
	"testing"
)

func TestUserDir(t *testing.T) {
	got := UserDir("/data", "alice")
	want := filepath.Join("/data", "users", "alice")
	if got != want {
		t.Fatalf("UserDir() = %q, want %q", got, want)
	}
}
