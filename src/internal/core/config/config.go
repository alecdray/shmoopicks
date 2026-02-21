package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Unable to load .env file")
	}
}

type Config struct {
	Port       string
	ConfigPath string
}

func LoadConfig() *Config {
	return &Config{
		Port: GetEnvOrDefault("PORT", "8080"),
	}
}

func (config *Config) ValidateConfig() error {
	return nil
}

func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
