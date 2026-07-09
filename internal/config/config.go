package config

import (
	"log"
	"os"

	"github.com/solargate/travka/internal/data"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Name string `mapstructure:"name" yaml:"name"`
		Port int    `mapstructure:"port" yaml:"port"`
		TLS  struct {
			Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
			Port     int    `mapstructure:"port" yaml:"port"`
			CertFile string `mapstructure:"cert_file" yaml:"cert_file"`
			KeyFile  string `mapstructure:"key_file" yaml:"key_file"`
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
		Location    string `mapstructure:"location" yaml:"location"`
		ResolvedDir string `mapstructure:"-" yaml:"-"`
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

	if Cfg.Auth.JWTTTLHours <= 0 {
		Cfg.Auth.JWTTTLHours = 24
	}
	if Cfg.Server.TLS.Enabled && Cfg.Server.TLS.Port <= 0 {
		Cfg.Server.TLS.Port = 8443
	}
	if Cfg.Federation.DeliveryWorkers <= 0 {
		Cfg.Federation.DeliveryWorkers = 2
	}
	if Cfg.Federation.DeliveryRetryMax <= 0 {
		Cfg.Federation.DeliveryRetryMax = 5
	}
	if Cfg.Data.Location == "" {
		Cfg.Data.Location = "data"
	}
	if Cfg.Auth.JWTSecret == "" {
		log.Fatalln("auth.jwt_secret must be set in config")
	}

	resolvedDir, err := data.ResolveDataDir(Cfg.Data.Location)
	if err != nil {
		log.Fatalf("Error resolving data location: %v", err)
	}
	if err := os.MkdirAll(resolvedDir, 0700); err != nil {
		log.Fatalf("Error creating data directory: %v", err)
	}
	Cfg.Data.ResolvedDir = resolvedDir
}
