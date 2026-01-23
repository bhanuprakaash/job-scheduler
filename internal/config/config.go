package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PG_DB_URL             string
	GRPC_PORT             string
	WORKERS_COUNT         int
	POLL_INTERVAL_SECONDS int
	RESEND_EMAIL_API_KEY  string
	RESEND_FROM_EMAIL     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dbUrl := os.Getenv("PG_DB_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}

	return &Config{
		PG_DB_URL:             dbUrl,
		GRPC_PORT:             getEnv("GRPC_PORT", "50052"),
		POLL_INTERVAL_SECONDS: getEnvAsInt("POLL_INTERVAL_SECONDS", 2),
		WORKERS_COUNT:         getEnvAsInt("WORKERS_COUNT", 5),
		RESEND_EMAIL_API_KEY:  getEnv("RESEND_EMAIL_API_KEY", ""),
		RESEND_FROM_EMAIL:     getEnv("RESEND_FROM_EMAIL", ""),
	}, nil
}

func getEnvAsInt(key string, fallback int) int {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("environment variable %s must be an integer", key)
		return fallback
	}
	return value
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
