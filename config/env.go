package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		slog.Warn("no .env file found, using system environment variables", "path", path)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
