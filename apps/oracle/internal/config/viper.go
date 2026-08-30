package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	_, err := os.Stat("./.env")
	if err != nil {
		log.Fatalf("failed to read env file: %v", err)
	}

	config := viper.New()

	config.SetConfigName(".env")
	config.SetConfigType("env")
	config.AddConfigPath("./")

	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("failed to read env: %v", err)
	}

	return config
}
