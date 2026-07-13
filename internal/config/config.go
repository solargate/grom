package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/solargate/grom/internal/data"
	"github.com/spf13/viper"
)

type TLSMode string

const (
	TLSModeOff      TLSMode = "off"
	TLSModeStatic   TLSMode = "static"
	TLSModeAutocert TLSMode = "autocert"
)

type Config struct {
	Server struct {
		Name string `mapstructure:"name" yaml:"name"`
		Port int    `mapstructure:"port" yaml:"port"`
		TLS  struct {
			Mode     string `mapstructure:"mode" yaml:"mode"`
			Enabled  bool   `mapstructure:"enabled" yaml:"enabled"` // legacy: true => static
			Port     int    `mapstructure:"port" yaml:"port"`
			CertFile string `mapstructure:"cert_file" yaml:"cert_file"`
			KeyFile  string `mapstructure:"key_file" yaml:"key_file"`
			Autocert struct {
				Email    string   `mapstructure:"email" yaml:"email"`
				CacheDir string   `mapstructure:"cache_dir" yaml:"cache_dir"`
				Domains  []string `mapstructure:"domains" yaml:"domains"`
			} `mapstructure:"autocert" yaml:"autocert"`
		} `mapstructure:"tls" yaml:"tls"`
	} `mapstructure:"server" yaml:"server"`
	Auth struct {
		JWTSecret   string `mapstructure:"jwt_secret" yaml:"jwt_secret"`
		JWTTTLHours int    `mapstructure:"jwt_ttl_hours" yaml:"jwt_ttl_hours"`
	} `mapstructure:"auth" yaml:"auth"`
	Federation struct {
		Enabled               bool   `mapstructure:"enabled" yaml:"enabled"`
		Domain                string `mapstructure:"domain" yaml:"domain"`
		AutoAcceptFollows     bool   `mapstructure:"auto_accept_follows" yaml:"auto_accept_follows"`
		DeliveryWorkers       int    `mapstructure:"delivery_workers" yaml:"delivery_workers"`
		DeliveryRetryMax      int    `mapstructure:"delivery_retry_max" yaml:"delivery_retry_max"`
		CACertFile            string `mapstructure:"ca_cert_file" yaml:"ca_cert_file"`
		TLSInsecureSkipVerify bool   `mapstructure:"tls_insecure_skip_verify" yaml:"tls_insecure_skip_verify"`
	} `mapstructure:"federation" yaml:"federation"`
	Data struct {
		Location        string `mapstructure:"location" yaml:"location"`
		TempDir         string `mapstructure:"temp_dir" yaml:"temp_dir"`
		ResolvedDir     string `mapstructure:"-" yaml:"-"`
		ResolvedTempDir string `mapstructure:"-" yaml:"-"`
	} `mapstructure:"data" yaml:"data"`
}

var Cfg Config

func GetConfig(configPath string) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		log.Fatalln("Error unmarshaling config")
	}

	if err := FinalizeConfig(&Cfg); err != nil {
		log.Fatal(err)
	}
}

func FinalizeConfig(cfg *Config) error {
	if cfg.Auth.JWTTTLHours <= 0 {
		cfg.Auth.JWTTTLHours = 24
	}
	if cfg.Federation.DeliveryWorkers <= 0 {
		cfg.Federation.DeliveryWorkers = 2
	}
	if cfg.Federation.DeliveryRetryMax <= 0 {
		cfg.Federation.DeliveryRetryMax = 5
	}
	if cfg.Data.Location == "" {
		cfg.Data.Location = "data"
	}
	if cfg.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret must be set in config")
	}

	mode := ResolveTLSMode(cfg)
	cfg.Server.TLS.Mode = string(mode)

	switch mode {
	case TLSModeOff:
		if cfg.Server.Port <= 0 {
			cfg.Server.Port = 8080
		}
	case TLSModeStatic:
		if cfg.Server.TLS.Port <= 0 {
			cfg.Server.TLS.Port = 8443
		}
		if cfg.Server.Port <= 0 {
			cfg.Server.Port = 8080
		}
	case TLSModeAutocert:
		if cfg.Server.TLS.Port <= 0 {
			cfg.Server.TLS.Port = 443
		}
		if cfg.Server.Port <= 0 {
			cfg.Server.Port = 8080
		}
	default:
		return fmt.Errorf("server.tls.mode must be one of off, static, autocert (got %q)", mode)
	}

	if cfg.Federation.Enabled && mode == TLSModeOff {
		return fmt.Errorf("federation.enabled requires server TLS (mode static or autocert)")
	}

	if mode == TLSModeStatic {
		if cfg.Server.TLS.CertFile == "" || cfg.Server.TLS.KeyFile == "" {
			return fmt.Errorf("server.tls.cert_file and server.tls.key_file are required when tls.mode is static")
		}
	}

	if mode == TLSModeAutocert {
		if cfg.Federation.Domain == "" {
			return fmt.Errorf("federation.domain is required when tls.mode is autocert")
		}
		if HasPort(cfg.Federation.Domain) {
			return fmt.Errorf("federation.domain must not include a port when tls.mode is autocert")
		}
		host := HostWithoutPort(cfg.Federation.Domain)
		if net.ParseIP(host) != nil {
			return fmt.Errorf("federation.domain must be a hostname, not an IP address, when tls.mode is autocert")
		}
		if len(cfg.Server.TLS.Autocert.Domains) == 0 {
			cfg.Server.TLS.Autocert.Domains = []string{host}
		}
		if cfg.Federation.TLSInsecureSkipVerify {
			log.Printf("warning: federation.tls_insecure_skip_verify is enabled with tls.mode autocert")
		}
	}

	resolvedDir, err := data.ResolveDataDir(cfg.Data.Location)
	if err != nil {
		return fmt.Errorf("resolve data location: %w", err)
	}
	if err := os.MkdirAll(resolvedDir, 0700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	cfg.Data.ResolvedDir = resolvedDir

	tempDir := strings.TrimSpace(cfg.Data.TempDir)
	if tempDir == "" {
		tempDir = "tmp"
	}
	resolvedTempDir, err := data.ResolveDataDir(tempDir)
	if err != nil {
		return fmt.Errorf("resolve data temp dir: %w", err)
	}
	if err := os.MkdirAll(resolvedTempDir, 0700); err != nil {
		return fmt.Errorf("create data temp directory: %w", err)
	}
	cfg.Data.ResolvedTempDir = resolvedTempDir

	if mode == TLSModeAutocert {
		if cfg.Server.TLS.Autocert.CacheDir == "" {
			cfg.Server.TLS.Autocert.CacheDir = filepath.Join(resolvedDir, "acme-cache")
		}
		if err := os.MkdirAll(cfg.Server.TLS.Autocert.CacheDir, 0700); err != nil {
			return fmt.Errorf("create acme cache directory: %w", err)
		}
	}

	return nil
}

func ResolveTLSMode(cfg *Config) TLSMode {
	mode := strings.TrimSpace(cfg.Server.TLS.Mode)
	if mode != "" {
		return TLSMode(mode)
	}
	if cfg.Server.TLS.Enabled {
		return TLSModeStatic
	}
	return TLSModeOff
}

func HostWithoutPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

func HasPort(host string) bool {
	_, _, err := net.SplitHostPort(host)
	return err == nil
}
