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
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dbUrl := os.Getenv("PG_DB_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}

	workersCount, err := getEnvAsInt("WORKERS_COUNT", 5)
	if err != nil {
		return nil, fmt.Errorf("invalid WORKERS_COUNT: %w", err)
	}

	pollInterval, err := getEnvAsInt("POLL_INTERVAL_SECONDS", 2)
	if err != nil {
		return nil, fmt.Errorf("invalid POLL_INTERVAL_SECONDS: %w", err)
	}

	return &Config{
		PG_DB_URL:             dbUrl,
		GRPC_PORT:             getEnv("GRPC_PORT", "50052"),
		POLL_INTERVAL_SECONDS: pollInterval,
		WORKERS_COUNT:         workersCount,
	}, nil
}

func getEnvAsInt(key string, fallback int) (int, error) {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return fallback, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("environment variable %s must be an integer", key)
	}
	return value, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
