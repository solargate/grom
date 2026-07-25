package config_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/solargate/grom/internal/config"
)

func baseCfg() config.Config {
	cfg := config.Config{}
	cfg.Auth.JWTSecret = "change-me-in-production-min-32-chars!!"
	cfg.Storage.Location = "data"
	return cfg
}

func TestFinalizeConfig_TLSOffDefaultsHTTPPort(t *testing.T) {
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

func TestFinalizeConfig_StaticTLS(t *testing.T) {
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

func TestFinalizeConfig_Autocert(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "autocert"
	cfg.Federation.Enabled = true
	cfg.Federation.Domain = "grom.example.com"

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Server.TLS.Port != 443 {
		t.Fatalf("tls port = %d, want 443", cfg.Server.TLS.Port)
	}
	if len(cfg.Server.TLS.Autocert.Domains) != 1 || cfg.Server.TLS.Autocert.Domains[0] != "grom.example.com" {
		t.Fatalf("domains = %v", cfg.Server.TLS.Autocert.Domains)
	}
	if cfg.Server.TLS.Autocert.CacheDir != "" {
		t.Fatalf("CacheDir = %q, want empty (YAML value preserved)", cfg.Server.TLS.Autocert.CacheDir)
	}
	wantCache := filepath.Join(cfg.Storage.ResolvedLocation, "acme-cache")
	if cfg.Server.TLS.Autocert.ResolvedCacheDir != wantCache {
		t.Fatalf("ResolvedCacheDir = %q, want %q", cfg.Server.TLS.Autocert.ResolvedCacheDir, wantCache)
	}
}

func TestFinalizeConfig_AutocertExplicitCacheDir(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "autocert"
	cfg.Federation.Enabled = true
	cfg.Federation.Domain = "grom.example.com"
	cfg.Server.TLS.Autocert.CacheDir = "custom-acme"

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Server.TLS.Autocert.CacheDir != "custom-acme" {
		t.Fatalf("CacheDir = %q, want custom-acme", cfg.Server.TLS.Autocert.CacheDir)
	}
	if !filepath.IsAbs(cfg.Server.TLS.Autocert.ResolvedCacheDir) {
		t.Fatalf("ResolvedCacheDir = %q, want absolute", cfg.Server.TLS.Autocert.ResolvedCacheDir)
	}
	if filepath.Base(cfg.Server.TLS.Autocert.ResolvedCacheDir) != "custom-acme" {
		t.Fatalf("ResolvedCacheDir = %q, want base custom-acme", cfg.Server.TLS.Autocert.ResolvedCacheDir)
	}
}

func TestFinalizeConfig_ResolvesTLSAndCAPaths(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "static"
	cfg.Server.TLS.CertFile = "tls/server.crt"
	cfg.Server.TLS.KeyFile = "tls/server.key"
	cfg.Federation.Enabled = true
	cfg.Federation.Domain = "192.168.1.251:8443"
	cfg.Federation.CACertFile = "tls/ca.crt"
	cfg.Federation.TLSInsecureSkipVerify = true

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(cfg.Server.TLS.ResolvedCertFile) {
		t.Fatalf("ResolvedCertFile = %q, want absolute", cfg.Server.TLS.ResolvedCertFile)
	}
	if !filepath.IsAbs(cfg.Server.TLS.ResolvedKeyFile) {
		t.Fatalf("ResolvedKeyFile = %q, want absolute", cfg.Server.TLS.ResolvedKeyFile)
	}
	if !filepath.IsAbs(cfg.Federation.ResolvedCACertFile) {
		t.Fatalf("ResolvedCACertFile = %q, want absolute", cfg.Federation.ResolvedCACertFile)
	}
	if !strings.HasSuffix(cfg.Server.TLS.ResolvedCertFile, filepath.Join("tls", "server.crt")) {
		t.Fatalf("ResolvedCertFile = %q, want suffix tls/server.crt", cfg.Server.TLS.ResolvedCertFile)
	}
	if !strings.HasSuffix(cfg.Federation.ResolvedCACertFile, filepath.Join("tls", "ca.crt")) {
		t.Fatalf("ResolvedCACertFile = %q, want suffix tls/ca.crt", cfg.Federation.ResolvedCACertFile)
	}
}

func TestFinalizeConfig_AbsoluteTLSPathsUnchanged(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "static"
	absCert := filepath.Join(t.TempDir(), "server.crt")
	absKey := filepath.Join(t.TempDir(), "server.key")
	absCA := filepath.Join(t.TempDir(), "ca.crt")
	cfg.Server.TLS.CertFile = absCert
	cfg.Server.TLS.KeyFile = absKey
	cfg.Federation.CACertFile = absCA
	cfg.Federation.Enabled = false

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Server.TLS.ResolvedCertFile != absCert {
		t.Fatalf("ResolvedCertFile = %q, want %q", cfg.Server.TLS.ResolvedCertFile, absCert)
	}
	if cfg.Server.TLS.ResolvedKeyFile != absKey {
		t.Fatalf("ResolvedKeyFile = %q, want %q", cfg.Server.TLS.ResolvedKeyFile, absKey)
	}
	if cfg.Federation.ResolvedCACertFile != absCA {
		t.Fatalf("ResolvedCACertFile = %q, want %q", cfg.Federation.ResolvedCACertFile, absCA)
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
	cfg.Federation.Domain = "grom.example.com:8443"

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

func TestFinalizeConfig_StorageTempDir(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "off"
	cfg.Storage.TempDir = "custom-tmp"

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Storage.ResolvedTempDir == "" {
		t.Fatal("ResolvedTempDir should be set")
	}
	if !filepath.IsAbs(cfg.Storage.ResolvedTempDir) {
		t.Fatalf("ResolvedTempDir = %q, want absolute path", cfg.Storage.ResolvedTempDir)
	}
}

func TestFinalizeConfig_StorageTempDirDefault(t *testing.T) {
	cfg := baseCfg()
	cfg.Server.TLS.Mode = "off"

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Storage.ResolvedTempDir == "" {
		t.Fatal("ResolvedTempDir should default to tmp")
	}
}

func TestFinalizeConfig_LegacyDataSection(t *testing.T) {
	cfg := config.Config{}
	cfg.Auth.JWTSecret = "change-me-in-production-min-32-chars!!"
	cfg.Server.TLS.Mode = "off"
	cfg.Data.Location = "legacy-data"
	cfg.Data.TempDir = "legacy-tmp"

	if err := config.FinalizeConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Storage.Location != "legacy-data" {
		t.Fatalf("Storage.Location = %q, want legacy-data", cfg.Storage.Location)
	}
	if cfg.Storage.TempDir != "legacy-tmp" {
		t.Fatalf("Storage.TempDir = %q, want legacy-tmp", cfg.Storage.TempDir)
	}
	if cfg.Storage.ResolvedLocation == "" {
		t.Fatal("ResolvedLocation should be set from legacy data section")
	}
}

func TestHostWithoutPort(t *testing.T) {
	if got := config.HostWithoutPort("192.168.1.251:8443"); got != "192.168.1.251" {
		t.Fatalf("HostWithoutPort = %q", got)
	}
	if got := config.HostWithoutPort("grom.example.com"); got != "grom.example.com" {
		t.Fatalf("HostWithoutPort = %q", got)
	}
}
