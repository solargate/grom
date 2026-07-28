package keys_test

import (
	"strings"
	"testing"

	"github.com/solargate/grom/internal/storage/keys"
)

func TestKeyHelpers(t *testing.T) {
	if got := keys.UserAvatar("alice"); !strings.Contains(got, "alice") {
		t.Fatalf("UserAvatar = %q", got)
	}
	if got := keys.WorkoutTrack("alice", "dir", "track.gpx"); !strings.HasSuffix(got, "track.gpx") {
		t.Fatalf("WorkoutTrack = %q", got)
	}
	if got := keys.UserActorKey("alice"); !strings.Contains(got, "alice") {
		t.Fatalf("UserActorKey = %q", got)
	}
	if got := keys.FederatedInboxTrack("viewer", "owner", "wid", "track.gpx"); !strings.Contains(got, "viewer") {
		t.Fatalf("FederatedInboxTrack = %q", got)
	}
	if got := keys.WorkoutSpeed("alice", "dir", keys.SpeedFileYAML); !strings.HasSuffix(got, "speed.yaml") {
		t.Fatalf("WorkoutSpeed = %q", got)
	}
	if got := keys.FederatedInboxSpeed("viewer", "owner", "wid", keys.SpeedFileJSON); !strings.HasSuffix(got, "wid_speed.json") {
		t.Fatalf("FederatedInboxSpeed = %q", got)
	}
}
