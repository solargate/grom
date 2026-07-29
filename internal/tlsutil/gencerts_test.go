package tlsutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solargate/grom/internal/tlsutil"
)

func TestGenerateCertsRequiresHost(t *testing.T) {
	err := tlsutil.GenerateCerts(tlsutil.GenOptions{OutDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected error without domain or ip")
	}
}

func TestGenerateCertsWithDomain(t *testing.T) {
	dir := t.TempDir()
	if err := tlsutil.GenerateCerts(tlsutil.GenOptions{
		Domain: "grom.test",
		OutDir: dir,
	}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"server.crt", "server.key", "ca.crt", "ca.key"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}

func TestGenerateCertsWithIP(t *testing.T) {
	dir := t.TempDir()
	if err := tlsutil.GenerateCerts(tlsutil.GenOptions{
		IP:     "127.0.0.1",
		OutDir: dir,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "server.crt")); err != nil {
		t.Fatal(err)
	}
}
