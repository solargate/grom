package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
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
}
