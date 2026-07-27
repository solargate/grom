package config

import (
	"fmt"
	"log"
	"log/slog"
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

type StorageDriver string

const (
	StorageDriverFile     StorageDriver = "file"
	StorageDriverBBolt    StorageDriver = "bbolt"
	StorageDriverPostgres StorageDriver = "postgres"
)

type StorageConfig struct {
	Driver           StorageDriver `mapstructure:"driver" yaml:"driver"`
	Location         string        `mapstructure:"location" yaml:"location"`
	TempDir          string        `mapstructure:"temp_dir" yaml:"temp_dir"`
	ResolvedLocation string        `mapstructure:"-" yaml:"-"`
	ResolvedTempDir  string        `mapstructure:"-" yaml:"-"`
	ResolvedBBoltPath string       `mapstructure:"-" yaml:"-"`
	BBolt struct {
		Path string `mapstructure:"path" yaml:"path"`
	} `mapstructure:"bbolt" yaml:"bbolt"`
	Postgres struct {
		DSN string `mapstructure:"dsn" yaml:"dsn"`
	} `mapstructure:"postgres" yaml:"postgres"`
}

type Config struct {
	Server struct {
		Name string `mapstructure:"name" yaml:"name"`
		Port int    `mapstructure:"port" yaml:"port"`
		TLS  struct {
			Mode             string `mapstructure:"mode" yaml:"mode"`
			Enabled          bool   `mapstructure:"enabled" yaml:"enabled"` // legacy: true => static
			Port             int    `mapstructure:"port" yaml:"port"`
			CertFile         string `mapstructure:"cert_file" yaml:"cert_file"`
			KeyFile          string `mapstructure:"key_file" yaml:"key_file"`
			ResolvedCertFile string `mapstructure:"-" yaml:"-"`
			ResolvedKeyFile  string `mapstructure:"-" yaml:"-"`
			Autocert struct {
				Email            string   `mapstructure:"email" yaml:"email"`
				CacheDir         string   `mapstructure:"cache_dir" yaml:"cache_dir"`
				ResolvedCacheDir string   `mapstructure:"-" yaml:"-"`
				Domains          []string `mapstructure:"domains" yaml:"domains"`
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
		ResolvedCACertFile    string `mapstructure:"-" yaml:"-"`
		TLSInsecureSkipVerify bool   `mapstructure:"tls_insecure_skip_verify" yaml:"tls_insecure_skip_verify"`
	} `mapstructure:"federation" yaml:"federation"`
	Storage StorageConfig `mapstructure:"storage" yaml:"storage"`
	// Data is a legacy alias for Storage (location/temp_dir only).
	Data StorageConfig `mapstructure:"data" yaml:"data"`
	Logging LoggingConfig `mapstructure:"logging" yaml:"logging"`
}

// LoggingConfig controls structured server logging (slog).
type LoggingConfig struct {
	// Level is debug, info, warn, or error. Default: info.
	Level string `mapstructure:"level" yaml:"level"`
	// Format is text or json. Default: json.
	Format string `mapstructure:"format" yaml:"format"`
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
	if err := finalizeLogging(&cfg.Logging); err != nil {
		return err
	}
	if cfg.Auth.JWTTTLHours <= 0 {
		cfg.Auth.JWTTTLHours = 24
	}
	if cfg.Federation.DeliveryWorkers <= 0 {
		cfg.Federation.DeliveryWorkers = 2
	}
	if cfg.Federation.DeliveryRetryMax <= 0 {
		cfg.Federation.DeliveryRetryMax = 5
	}
	if cfg.Storage.Location == "" && cfg.Data.Location != "" {
		cfg.Storage.Location = cfg.Data.Location
	}
	if cfg.Storage.TempDir == "" && cfg.Data.TempDir != "" {
		cfg.Storage.TempDir = cfg.Data.TempDir
	}
	if cfg.Storage.Location == "" {
		cfg.Storage.Location = "data"
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
			slog.Warn("federation.tls_insecure_skip_verify is enabled with tls.mode autocert")
		}
	}

	resolvedDir, err := data.ResolveDataDir(cfg.Storage.Location)
	if err != nil {
		return fmt.Errorf("resolve storage location: %w", err)
	}
	if err := os.MkdirAll(resolvedDir, 0700); err != nil {
		return fmt.Errorf("create storage directory: %w", err)
	}
	cfg.Storage.ResolvedLocation = resolvedDir

	bboltPath := strings.TrimSpace(cfg.Storage.BBolt.Path)
	if bboltPath == "" {
		cfg.Storage.ResolvedBBoltPath = filepath.Join(resolvedDir, "grom.db")
	} else {
		resolvedBBolt, err := data.ResolveDataDir(bboltPath)
		if err != nil {
			return fmt.Errorf("resolve storage.bbolt.path: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(resolvedBBolt), 0700); err != nil {
			return fmt.Errorf("create bbolt parent directory: %w", err)
		}
		cfg.Storage.ResolvedBBoltPath = resolvedBBolt
	}

	tempDir := strings.TrimSpace(cfg.Storage.TempDir)
	if tempDir == "" {
		tempDir = "tmp"
	}
	resolvedTempDir, err := data.ResolveDataDir(tempDir)
	if err != nil {
		return fmt.Errorf("resolve storage temp dir: %w", err)
	}
	if err := os.MkdirAll(resolvedTempDir, 0700); err != nil {
		return fmt.Errorf("create storage temp directory: %w", err)
	}
	cfg.Storage.ResolvedTempDir = resolvedTempDir

	// Keep legacy Data fields in sync for any remaining references during migration.
	cfg.Data.ResolvedLocation = cfg.Storage.ResolvedLocation
	cfg.Data.ResolvedTempDir = cfg.Storage.ResolvedTempDir
	cfg.Data.Location = cfg.Storage.Location
	cfg.Data.TempDir = cfg.Storage.TempDir

	if mode == TLSModeStatic {
		resolvedCert, err := data.ResolveDataDir(cfg.Server.TLS.CertFile)
		if err != nil {
			return fmt.Errorf("resolve server.tls.cert_file: %w", err)
		}
		resolvedKey, err := data.ResolveDataDir(cfg.Server.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("resolve server.tls.key_file: %w", err)
		}
		cfg.Server.TLS.ResolvedCertFile = resolvedCert
		cfg.Server.TLS.ResolvedKeyFile = resolvedKey
	}

	if ca := strings.TrimSpace(cfg.Federation.CACertFile); ca != "" {
		resolvedCA, err := data.ResolveDataDir(ca)
		if err != nil {
			return fmt.Errorf("resolve federation.ca_cert_file: %w", err)
		}
		cfg.Federation.ResolvedCACertFile = resolvedCA
	}

	if mode == TLSModeAutocert {
		cacheDir := strings.TrimSpace(cfg.Server.TLS.Autocert.CacheDir)
		if cacheDir == "" {
			cacheDir = "acme-cache"
		}
		resolvedCacheDir, err := data.ResolveDataDir(cacheDir)
		if err != nil {
			return fmt.Errorf("resolve server.tls.autocert.cache_dir: %w", err)
		}
		if err := os.MkdirAll(resolvedCacheDir, 0700); err != nil {
			return fmt.Errorf("create acme cache directory: %w", err)
		}
		cfg.Server.TLS.Autocert.ResolvedCacheDir = resolvedCacheDir
	}

	return nil
}

func finalizeLogging(cfg *LoggingConfig) error {
	level := strings.ToLower(strings.TrimSpace(cfg.Level))
	if level == "" {
		level = "info"
	}
	switch level {
	case "debug", "info", "warn", "error":
		cfg.Level = level
	case "warning":
		cfg.Level = "warn"
	default:
		return fmt.Errorf("logging.level must be one of debug, info, warn, error (got %q)", cfg.Level)
	}

	format := strings.ToLower(strings.TrimSpace(cfg.Format))
	if format == "" {
		format = "json"
	}
	switch format {
	case "text", "json":
		cfg.Format = format
	default:
		return fmt.Errorf("logging.format must be text or json (got %q)", cfg.Format)
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
