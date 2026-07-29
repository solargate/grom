package federation

import "testing"

func TestOwnerKeyFromHandle(t *testing.T) {
	tests := []struct {
		handle string
		want   string
	}{
		{"alice@example.com", "alice_example.com"},
		{"bob@192.168.1.1:8445", "bob_192.168.1.1_8445"},
		{"carol@host/path", "carol_host_path"},
	}
	for _, tc := range tests {
		if got := OwnerKeyFromHandle(tc.handle); got != tc.want {
			t.Fatalf("OwnerKeyFromHandle(%q) = %q, want %q", tc.handle, got, tc.want)
		}
	}
}

func TestOwnerHandleFromKeyRoundTripForAt(t *testing.T) {
	handle := "alice@example.com"
	key := OwnerKeyFromHandle(handle)
	if got := OwnerHandleFromKey(key); got != handle {
		t.Fatalf("OwnerHandleFromKey(%q) = %q, want %q", key, got, handle)
	}
}

func TestOwnerHandleFromKeyIsLossyForPort(t *testing.T) {
	handle := "bob@host:8445"
	key := OwnerKeyFromHandle(handle)
	got := OwnerHandleFromKey(key)
	if got == handle {
		t.Fatalf("expected lossy reverse for %q, got identical", handle)
	}
	if got != "bob@host@8445" {
		t.Fatalf("OwnerHandleFromKey(%q) = %q", key, got)
	}
}

func TestOwnerNicknameFromKey(t *testing.T) {
	if got := OwnerNicknameFromKey(OwnerKeyFromHandle("alice@example.com")); got != "alice" {
		t.Fatalf("nickname = %q", got)
	}
	if got := OwnerNicknameFromKey("solo"); got != "solo" {
		t.Fatalf("solo nickname = %q", got)
	}
}
