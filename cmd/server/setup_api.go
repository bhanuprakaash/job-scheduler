package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/bhanuprakaash/job-scheduler/internal/api"
	"github.com/bhanuprakaash/job-scheduler/internal/config"
	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
	pb "github.com/bhanuprakaash/job-scheduler/proto"
	"google.golang.org/grpc"
)

func runGRPCServer(cfg *config.Config, db *store.Store) {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC_PORT))
	if err != nil {
		logger.Fatal("Failed to listen", "error", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterJobSchedulerServer(grpcServer, api.NewServer(db))

	logger.Info("gRPC server listening", "port", cfg.GRPC_PORT)

	// Run in goroutine so main can listen for signals
	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			logger.Fatal("Failed to serve", "error", err)
		}
	}()

	// Graceful Shutdown Logic
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down gRPC server")
	grpcServer.GracefulStop()
	logger.Info("Server stopped")
}
