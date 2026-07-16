package internal

import (
	"errors"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
}

func Load() Config {
	err := godotenv.Load("config.env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("failed to load config.env", "error", err)
		os.Exit(1)
	}

	return Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		DatabaseURL: requireEnv("DATABASE_URL"),
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		slog.Error("required environment variable is empty", "key", key)
		os.Exit(1)
	}

	return value
}
