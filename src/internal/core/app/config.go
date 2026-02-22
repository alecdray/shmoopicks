package app

import (
	"fmt"
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
	Port                string
	DbPath              string
	JwtSecret           string
	Host                string
	StateCode           string
	SpotifyClientId     string
	SpotifyClientSecret string
}

func LoadConfig() *Config {
	port := GetEnvOrDefault("PORT", "8080")
	host := GetEnvOrDefault("Host", fmt.Sprintf("http://127.0.0.1:%s", port))
	return &Config{
		Port:                port,
		DbPath:              GetEnvOrDefault("DB_PATH", "./tmp/db.sql"),
		JwtSecret:           GetEnvOrDefault("JWT_SECRET", "secret"),
		Host:                host,
		StateCode:           GetEnvOrDefault("STATE_CODE", "state"),
		SpotifyClientId:     GetEnvOrDefault("SPOTIFY_CLIENT_ID", ""),
		SpotifyClientSecret: GetEnvOrDefault("SPOTIFY_CLIENT_SECRET", ""),
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
