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
		Name         string           `mapstructure:"name" yaml:"name"`
		Port         int              `mapstructure:"port" yaml:"port"`
		Registration RegistrationMode `mapstructure:"registration" yaml:"registration"`
		TLS          struct {
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
		Reset       struct {
			// PublicBaseURL is the instance base URL used in password-reset email links (no trailing slash).
			PublicBaseURL string `mapstructure:"public_base_url" yaml:"public_base_url"`
			// TokenTTLMinutes is how long a reset token remains valid. Default: 60.
			TokenTTLMinutes int `mapstructure:"token_ttl_minutes" yaml:"token_ttl_minutes"`
		} `mapstructure:"reset" yaml:"reset"`
		Captcha CaptchaConfig `mapstructure:"captcha" yaml:"captcha"`
	} `mapstructure:"auth" yaml:"auth"`
	Mailer MailerConfig `mapstructure:"mailer" yaml:"mailer"`
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

// RegistrationMode controls whether new users can sign up.
type RegistrationMode string

const (
	RegistrationOpen   RegistrationMode = "open"
	RegistrationClosed RegistrationMode = "closed"
	RegistrationInvite RegistrationMode = "invite"
)

// MailerDriver selects how outbound email is delivered.
type MailerDriver string

const (
	MailerDriverOff  MailerDriver = "off"
	MailerDriverLog  MailerDriver = "log"
	MailerDriverSMTP MailerDriver = "smtp"
)

// MailerEncryption selects SMTP transport security.
type MailerEncryption string

const (
	MailerEncryptionSTARTTLS MailerEncryption = "starttls"
	MailerEncryptionTLS      MailerEncryption = "tls"
	MailerEncryptionNone     MailerEncryption = "none"
)

// MailerConfig controls outbound email (password reset and future notifications).
type MailerConfig struct {
	// Driver is off, log, or smtp. Default: off.
	Driver string `mapstructure:"driver" yaml:"driver"`
	// From is the sender address, e.g. "Grom <noreply@example.com>".
	From string `mapstructure:"from" yaml:"from"`
	SMTP struct {
		Host       string `mapstructure:"host" yaml:"host"`
		Port       int    `mapstructure:"port" yaml:"port"`
		Username   string `mapstructure:"username" yaml:"username"`
		Password   string `mapstructure:"password" yaml:"password"`
		Encryption string `mapstructure:"encryption" yaml:"encryption"` // starttls | tls | none
	} `mapstructure:"smtp" yaml:"smtp"`
}

// MailerEnabled reports whether outbound email delivery is configured.
func (c *Config) MailerEnabled() bool {
	d := strings.ToLower(strings.TrimSpace(c.Mailer.Driver))
	return d == string(MailerDriverLog) || d == string(MailerDriverSMTP)
}

// PasswordResetEnabled reports whether password reset via email is available.
func (c *Config) PasswordResetEnabled() bool {
	return c.MailerEnabled() && strings.TrimSpace(c.Auth.Reset.PublicBaseURL) != ""
}

// CaptchaConfig controls optional ALTCHA proof-of-work protection on auth forms.
type CaptchaConfig struct {
	// Enabled turns on captcha for register, login, and password forgot. Default: false.
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`
	// HMACSecret signs challenges. Optional; when empty, auth.jwt_secret is used.
	HMACSecret string `mapstructure:"hmac_secret" yaml:"hmac_secret"`
	// Cost is the PBKDF2 iteration count for PoW. Default: 1000.
	Cost int `mapstructure:"cost" yaml:"cost"`
	// ExpiresSeconds is challenge lifetime. Default: 300.
	ExpiresSeconds int `mapstructure:"expires_seconds" yaml:"expires_seconds"`
}

// CaptchaEnabled reports whether ALTCHA captcha is required on auth endpoints.
func (c *Config) CaptchaEnabled() bool {
	return c.Auth.Captcha.Enabled
}

// RegistrationAllowed reports whether new user registration is open.
func (c *Config) RegistrationAllowed() bool {
	return c.Server.Registration == RegistrationOpen
}

// CaptchaHMACSecret returns the HMAC key used to sign/verify challenges.
func (c *Config) CaptchaHMACSecret() string {
	if s := strings.TrimSpace(c.Auth.Captcha.HMACSecret); s != "" {
		return s
	}
	return c.Auth.JWTSecret
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
	if cfg.Auth.Reset.TokenTTLMinutes <= 0 {
		cfg.Auth.Reset.TokenTTLMinutes = 60
	}
	cfg.Auth.Reset.PublicBaseURL = strings.TrimRight(strings.TrimSpace(cfg.Auth.Reset.PublicBaseURL), "/")
	if cfg.Auth.Captcha.Cost <= 0 {
		cfg.Auth.Captcha.Cost = 1000
	}
	if cfg.Auth.Captcha.ExpiresSeconds <= 0 {
		cfg.Auth.Captcha.ExpiresSeconds = 300
	}
	if err := finalizeMailer(&cfg.Mailer, cfg.Auth.Reset.PublicBaseURL); err != nil {
		return err
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
	switch cfg.Server.Registration {
	case "":
		cfg.Server.Registration = RegistrationOpen
	case RegistrationOpen, RegistrationClosed, RegistrationInvite:
		// valid
	default:
		return fmt.Errorf("server.registration must be one of open, closed, invite (got %q)", cfg.Server.Registration)
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

func finalizeMailer(cfg *MailerConfig, publicBaseURL string) error {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		driver = string(MailerDriverOff)
	}
	switch MailerDriver(driver) {
	case MailerDriverOff, MailerDriverLog, MailerDriverSMTP:
		cfg.Driver = driver
	default:
		return fmt.Errorf("mailer.driver must be one of off, log, smtp (got %q)", cfg.Driver)
	}

	if MailerDriver(driver) == MailerDriverOff {
		return nil
	}

	cfg.From = strings.TrimSpace(cfg.From)
	if cfg.From == "" {
		return fmt.Errorf("mailer.from is required when mailer.driver is %s", driver)
	}
	if publicBaseURL == "" {
		return fmt.Errorf("auth.reset.public_base_url is required when mailer.driver is %s", driver)
	}

	if MailerDriver(driver) != MailerDriverSMTP {
		return nil
	}

	cfg.SMTP.Host = strings.TrimSpace(cfg.SMTP.Host)
	if cfg.SMTP.Host == "" {
		return fmt.Errorf("mailer.smtp.host is required when mailer.driver is smtp")
	}
	if cfg.SMTP.Port <= 0 {
		return fmt.Errorf("mailer.smtp.port is required when mailer.driver is smtp")
	}

	enc := strings.ToLower(strings.TrimSpace(cfg.SMTP.Encryption))
	if enc == "" {
		if cfg.SMTP.Port == 465 {
			enc = string(MailerEncryptionTLS)
		} else {
			enc = string(MailerEncryptionSTARTTLS)
		}
	}
	switch MailerEncryption(enc) {
	case MailerEncryptionSTARTTLS, MailerEncryptionTLS, MailerEncryptionNone:
		cfg.SMTP.Encryption = enc
	default:
		return fmt.Errorf("mailer.smtp.encryption must be one of starttls, tls, none (got %q)", cfg.SMTP.Encryption)
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
