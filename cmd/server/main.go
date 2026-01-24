package main

import (
	"context"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/config"
	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
	"github.com/bhanuprakaash/job-scheduler/internal/worker"
)

func main() {
	logger.Init()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load the config", "error", err)
	}

	// db connection
	db, err := store.NewStore(ctx, cfg.PG_DB_URL)
	if err != nil {
		logger.Fatal("Failed to ping the store", "error", err)
	}
	defer db.Close()
	logger.Info("Connected to Database")

	// job registry
	jobRegistry, err := setupJobRegistry(cfg)
	if err != nil {
		logger.Fatal("Failed to setup job registry", "error", err)
	}

	// worker pool
	workerPool := worker.NewPool(db, jobRegistry, cfg.WORKERS_COUNT, time.Duration(cfg.POLL_INTERVAL_SECONDS))
	workerPool.Start(ctx)
	defer workerPool.Stop()

	// grpc connection
	runGRPCServer(cfg, db)

}
