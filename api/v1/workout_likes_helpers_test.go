package v1

import "testing"

func TestRemoteWorkoutObjectID(t *testing.T) {
	tests := []struct {
		handle, nickname, workoutID, want string
	}{
		{"bob@remote.test", "bob", "38472901", "https://remote.test/users/bob/workouts/38472901"},
		{"bob@192.168.1.1:8443", "bob", "abc", "https://192.168.1.1:8443/users/bob/workouts/abc"},
		{"@onlyhost", "bob", "abc", "https://onlyhost/users/bob/workouts/abc"},
		{"nodomain", "bob", "abc", ""},
		{"", "bob", "abc", ""},
	}
	for _, tt := range tests {
		got := remoteWorkoutObjectID(tt.handle, tt.nickname, tt.workoutID)
		if got != tt.want {
			t.Fatalf("remoteWorkoutObjectID(%q,%q,%q) = %q, want %q",
				tt.handle, tt.nickname, tt.workoutID, got, tt.want)
		}
	}
}
