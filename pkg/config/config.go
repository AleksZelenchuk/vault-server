package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	DatabaseURL    string
	VaultMasterKey string
}

// LoadConfig from os to local struct for farther usage
func LoadConfig() *Config {
	Init()
	cfg := &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		VaultMasterKey: os.Getenv("VAULT_MASTER_KEY"),
	}

	if cfg.DatabaseURL == "" || cfg.VaultMasterKey == "" {
		log.Fatal("Missing one or more required environment variables")
	}

	return cfg
}

// Init will load data from .env file to os
func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
