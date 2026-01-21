package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PG_DB_URL string
	GRPC_PORT string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dbUrl := os.Getenv("PG_DB_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if dbUrl == "" {
		return nil, fmt.Errorf("GRPC_PORT is required")
	}

	return &Config{
		PG_DB_URL: dbUrl,
		GRPC_PORT: grpcPort,
	}, nil
}
