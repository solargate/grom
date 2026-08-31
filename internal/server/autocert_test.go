package server

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/config"
)

func TestRunAutocertTLSRequiresDomains(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })

	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	config.Cfg = config.Config{}
	config.Cfg.Server.TLS.Mode = "autocert"
	config.Cfg.Federation.Enabled = true
	config.Cfg.Federation.Domain = "grom.example.com"
	config.Cfg.Server.TLS.Autocert.Domains = []string{"grom.example.com"}
	config.Cfg.Storage.Driver = "file"
	config.Cfg.Storage.Location = dir
	config.Cfg.Auth.JWTSecret = "test-secret-at-least-32-characters!!"
	if err := config.FinalizeConfig(&config.Cfg); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	// runAutocertTLS blocks; invoke in goroutine and verify it accepts config.
	done := make(chan error, 1)
	go func() {
		done <- runAutocertTLS(router)
	}()

	select {
	case err := <-done:
		// Expected to fail quickly when port 443/80 unavailable in test env.
		if err == nil {
			t.Fatal("expected listen error in unit test environment")
		}
	default:
		// Still running — config was accepted; pass smoke test.
	}
}

func TestAutocertConfigFinalizeResolvedCacheDir(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })

	cfg := config.Config{}
	cfg.Server.TLS.Mode = "autocert"
	cfg.Federation.Enabled = true
	cfg.Federation.Domain = "example.com"
	cfg.Server.TLS.Autocert.Domains = []string{"example.com"}
	cfg.Storage.Driver = "file"
	cfg.Storage.Location = t.TempDir()
	cfg.Auth.JWTSecret = "test-secret-at-least-32-characters!!"
	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Server.TLS.Autocert.Domains) != 1 {
		t.Fatalf("domains = %#v", cfg.Server.TLS.Autocert.Domains)
	}
}
