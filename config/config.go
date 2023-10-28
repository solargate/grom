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

func GetConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("Error reading config")
	}

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		log.Fatalln("Error unmarshaling config")
	}
}
