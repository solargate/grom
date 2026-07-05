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
	} `mapstructure:"server" yaml:"server"`
	Auth struct {
		JWTSecret   string `mapstructure:"jwt_secret" yaml:"jwt_secret"`
		JWTTTLHours int    `mapstructure:"jwt_ttl_hours" yaml:"jwt_ttl_hours"`
	} `mapstructure:"auth" yaml:"auth"`
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
