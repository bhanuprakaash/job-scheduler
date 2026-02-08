package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APP_ENV               string
	PG_DB_URL             string
	GRPC_PORT             string
	GRPC_HOST             string
	WORKERS_COUNT         int
	POLL_INTERVAL_SECONDS int
	HTTP_PORT             string
	METRICS_PORT          string

	// email
	RESEND_EMAIL_API_KEY string
	RESEND_FROM_EMAIL    string

	// minio (object upload)

	MINIO_ID       string
	MINIO_SECRET   string
	MINIO_ENDPOINT string
	MINIO_BUCKET   string
	MINIO_USE_SSL  bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		APP_ENV:               getEnv("APP_ENV", "development"),
		PG_DB_URL:             getEnv("PG_DB_URL", ""),
		GRPC_PORT:             getEnv("GRPC_PORT", "50052"),
		GRPC_HOST:             getEnv("GRPC_HOST", "localhost"),
		POLL_INTERVAL_SECONDS: getEnvAsInt("POLL_INTERVAL_SECONDS", 2),
		WORKERS_COUNT:         getEnvAsInt("WORKERS_COUNT", 5),
		HTTP_PORT:             getEnv("HTTP_PORT", "8080"),
		METRICS_PORT:          getEnv("METRICS_PORT", "9090"),

		RESEND_EMAIL_API_KEY: getEnv("RESEND_EMAIL_API_KEY", ""),
		RESEND_FROM_EMAIL:    getEnv("RESEND_FROM_EMAIL", ""),

		MINIO_ID:       getEnv("MINIO_ID", "minioadmin"),
		MINIO_SECRET:   getEnv("MINIO_SECRET", "minioadmin"),
		MINIO_ENDPOINT: getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MINIO_BUCKET:   getEnv("MINIO_BUCKET", "job-resize-images"),
		MINIO_USE_SSL:  getEnvAsBool("MINIO_USE_SSL", false),
	}

	if cfg.PG_DB_URL == "" {
		return nil, fmt.Errorf("PG_DB_URL is required")
	}

	if cfg.APP_ENV == "production" {
		if cfg.MINIO_ID == "minioadmin" || cfg.MINIO_SECRET == "minioadmin" {
			return nil, fmt.Errorf("CRITICAL: Cannot use default MinIO credentials in production. Set MINIO_ID and MINIO_SECRET")
		}
		if cfg.RESEND_EMAIL_API_KEY == "" {
			return nil, fmt.Errorf("CRITICAL: RESEND_EMAIL_API_KEY is required in production")
		}
	}

	return cfg, nil
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

func getEnvAsBool(key string, fallback bool) bool {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		fmt.Printf("environment variable %s must be a boolean\n", key)
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
