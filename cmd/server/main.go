package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/api"
	"github.com/bhanuprakaash/job-scheduler/internal/catalog/notifications"
	"github.com/bhanuprakaash/job-scheduler/internal/config"
	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
	"github.com/bhanuprakaash/job-scheduler/internal/worker"
	pb "github.com/bhanuprakaash/job-scheduler/proto"
	"google.golang.org/grpc"
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
	jobRegistry := worker.NewRegistry()
	jobRegistry.Register("dummy", &notifications.DummyJob{})

	// worker pool
	workerPool := worker.NewPool(db, jobRegistry, cfg.WORKERS_COUNT, time.Duration(cfg.POLL_INTERVAL_SECONDS))
	workerPool.Start(ctx)
	defer workerPool.Stop()

	// grpc connection
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC_PORT))
	if err != nil {
		logger.Fatal("Failed to listen", "error", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterJobSchedulerServer(grpcServer, api.NewServer(db))

	logger.Info("gRPC server listening", "port", cfg.GRPC_PORT)

	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			logger.Fatal("Failed to serve", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down gRPC server")
	grpcServer.GracefulStop()

	logger.Info("Server stopped")

}
