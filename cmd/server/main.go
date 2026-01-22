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
	"github.com/bhanuprakaash/job-scheduler/internal/config"
	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
	"github.com/bhanuprakaash/job-scheduler/internal/worker"
	pb "github.com/bhanuprakaash/job-scheduler/proto"
	"google.golang.org/grpc"
)

const (
	numWorkers   = 3
	pollInterval = 2 * time.Second
)

func main() {
	logger.Init()
	ctx := context.Background()

	config, err := config.Load()
	if err != nil {
		logger.Error("Failed to load the config", "error", err)
	}

	// db connection
	db, err := store.NewStore(ctx, config.PG_DB_URL)
	if err != nil {
		logger.Error("Failed to ping the store", "error", err)
	}
	defer db.Close()
	logger.Info("Connected to Database")

	// worker pool
	workerPool := worker.NewPool(db, numWorkers, pollInterval)
	workerPool.Start(ctx)
	defer workerPool.Stop()

	// grpc connection
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPC_PORT))
	if err != nil {
		logger.Error("Failed to listen", "error", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterJobSchedulerServer(grpcServer, api.NewServer(db))

	logger.Info("gRPC server listening", "port", config.GRPC_PORT)

	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			logger.Error("Failed to serve", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down gRPC server")
	grpcServer.GracefulStop()

	logger.Info("Server stopped")

}
