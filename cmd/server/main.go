package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/config"
	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
	"github.com/bhanuprakaash/job-scheduler/internal/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.Init()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	jobRegistry, err := setupJobRegistry(cfg, db)
	if err != nil {
		logger.Fatal("Failed to setup job registry", "error", err)
	}

	// worker pool
	workerPool := worker.NewPool(db, jobRegistry, cfg.WORKERS_COUNT, time.Duration(cfg.POLL_INTERVAL_SECONDS)*time.Second)
	workerPool.Start(ctx)
	defer workerPool.Stop()

	var wg sync.WaitGroup

	// grpc server
	wg.Add(1)
	go func() {
		defer wg.Done()
		runGRPCServer(ctx, cfg, db)
	}()

	// http gateway
	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Millisecond)
		runHTTPServer(ctx, cfg)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Metrics server starting", "port", "9090")

		http.Handle("/metrics", promhttp.Handler())

		if err := http.ListenAndServe(":9090", nil); err != nil {
			logger.Error("Metrics server failed", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Interrupt received, shutting down...")
	cancel()

	wg.Wait()
	logger.Info("All services stopped. Bye!")

}
