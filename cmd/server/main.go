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

	appCtx := context.Background()
	serverCtx, serverCancel := context.WithCancel(appCtx)
	defer serverCancel()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load the config", "error", err)
	}

	// db connection
	db, err := store.NewStore(appCtx, cfg.PG_DB_URL)
	if err != nil {
		logger.Fatal("Failed to ping the store", "error", err)
	}
	logger.Info("Connected to Database")

	// job registry
	jobRegistry, err := setupJobRegistry(cfg, db)
	if err != nil {
		logger.Fatal("Failed to setup job registry", "error", err)
	}

	// worker pool
	workerPool := worker.NewPool(db, jobRegistry, cfg.WORKERS_COUNT, time.Duration(cfg.POLL_INTERVAL_SECONDS)*time.Second)
	workerPool.Start(serverCtx)

	var wg sync.WaitGroup

	// grpc server
	wg.Add(1)
	go func() {
		defer wg.Done()
		runGRPCServer(serverCtx, cfg, db)
	}()

	// http gateway
	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Millisecond)
		runHTTPServer(serverCtx, cfg)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		srv := &http.Server{Addr: ":9090", Handler: mux}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("Metrics server error", "error", err)
			}
		}()

		<-serverCtx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Interrupt received, shutting down...")
	serverCancel()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Network services stopped")
	case <-time.After(5 * time.Second):
		logger.Info("Network services timed out shutting down")
	}

	logger.Info("Network services stopped")

	logger.Info("Draining worker pool...")
	workerPool.Stop()
	logger.Info("Worker pool drained")

	db.Close()
	logger.Info("Database closed")
	logger.Info("Bye!")

}
