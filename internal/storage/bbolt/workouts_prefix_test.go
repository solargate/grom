package bbolt

import (
	"bytes"
	"testing"
)

func TestIncrementPrefix(t *testing.T) {
	got := incrementPrefix([]byte("alice/"))
	if !bytes.Equal(got, []byte("alice0")) {
		t.Fatalf("incrementPrefix(alice/) = %q", got)
	}
	if incrementPrefix([]byte{0xFF, 0xFF}) != nil {
		t.Fatal("expected nil when prefix cannot increment")
	}
	got = incrementPrefix([]byte{0x00, 0xFF})
	if !bytes.Equal(got, []byte{0x01}) {
		t.Fatalf("carry = %v", got)
	}
}
