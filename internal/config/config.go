package config

import (
	"log"

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
	Users struct {
		File string `mapstructure:"file" yaml:"file"`
	} `mapstructure:"users" yaml:"users"`
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
	if Cfg.Users.File == "" {
		Cfg.Users.File = "./users.yaml"
	}
	if Cfg.Auth.JWTSecret == "" {
		log.Fatalln("auth.jwt_secret must be set in config")
	}
}
