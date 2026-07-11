package config_test

import (
	"path/filepath"
	"testing"

	"github.com/solargate/travka/internal/config"
)

func baseCfg() config.Config {
	cfg := config.Config{}
	cfg.Auth.JWTSecret = "change-me-in-production-min-32-chars!!"
	cfg.Data.Location = "data"
	return cfg
}

func TestFinalizeConfig_DevNoTLS(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "off"
	cfg.Federation.Enabled = false

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if config.ResolveTLSMode(&cfg) != config.TLSModeOff {
		t.Fatalf("mode = %q, want off", cfg.Server.TLS.Mode)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("port = %d, want 8080", cfg.Server.Port)
	}
}

func TestFinalizeConfig_DevStaticTLS(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "static"
	cfg.Server.TLS.Port = 8443
	cfg.Server.TLS.CertFile = "tls/server.crt"
	cfg.Server.TLS.KeyFile = "tls/server.key"
	cfg.Federation.Enabled = true
	cfg.Federation.Domain = "192.168.1.251:8443"
	cfg.Federation.TLSInsecureSkipVerify = true

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if config.ResolveTLSMode(&cfg) != config.TLSModeStatic {
		t.Fatalf("mode = %q, want static", cfg.Server.TLS.Mode)
	}
}

func TestFinalizeConfig_ProdNoTLS(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "off"
	cfg.Federation.Enabled = false

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
}

func TestFinalizeConfig_ProdAutocert(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "autocert"
	cfg.Federation.Enabled = true
	cfg.Federation.Domain = "travka.example.com"

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Server.TLS.Port != 443 {
		t.Fatalf("tls port = %d, want 443", cfg.Server.TLS.Port)
	}
	if len(cfg.Server.TLS.Autocert.Domains) != 1 || cfg.Server.TLS.Autocert.Domains[0] != "travka.example.com" {
		t.Fatalf("domains = %v", cfg.Server.TLS.Autocert.Domains)
	}
	wantCache := filepath.Join(cfg.Data.ResolvedDir, "acme-cache")
	if cfg.Server.TLS.Autocert.CacheDir != wantCache {
		t.Fatalf("cache_dir = %q, want %q", cfg.Server.TLS.Autocert.CacheDir, wantCache)
	}
}

func TestFinalizeConfig_FederationRequiresTLS(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "off"
	cfg.Federation.Enabled = true

	if err := config.FinalizeConfig(&cfg); err == nil {
		t.Fatal("expected error when federation enabled without TLS")
	}
}

func TestFinalizeConfig_StaticRequiresCertFiles(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "static"

	if err := config.FinalizeConfig(&cfg); err == nil {
		t.Fatal("expected error when static mode without cert files")
	}
}

func TestFinalizeConfig_AutocertRejectsIPDomain(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "autocert"
	cfg.Federation.Enabled = true
	cfg.Federation.Domain = "192.168.1.1"

	if err := config.FinalizeConfig(&cfg); err == nil {
		t.Fatal("expected error for IP domain with autocert")
	}
}

func TestFinalizeConfig_AutocertRejectsPortInDomain(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "autocert"
	cfg.Federation.Enabled = true
	cfg.Federation.Domain = "travka.example.com:8443"

	if err := config.FinalizeConfig(&cfg); err == nil {
		t.Fatal("expected error for domain with port and autocert")
	}
}

func TestResolveTLSMode_LegacyEnabled(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Enabled = true
	cfg.Server.TLS.CertFile = "tls/server.crt"
	cfg.Server.TLS.KeyFile = "tls/server.key"

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if config.ResolveTLSMode(&cfg) != config.TLSModeStatic {
		t.Fatalf("mode = %q, want static from legacy enabled flag", cfg.Server.TLS.Mode)
	}
}

func TestHostWithoutPort(t *testing.T) {
	if got := config.HostWithoutPort("192.168.1.251:8443"); got != "192.168.1.251" {
		t.Fatalf("HostWithoutPort = %q", got)
	}
	if got := config.HostWithoutPort("travka.example.com"); got != "travka.example.com" {
		t.Fatalf("HostWithoutPort = %q", got)
	}
}
