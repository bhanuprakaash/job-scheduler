package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/api"
	"github.com/bhanuprakaash/job-scheduler/internal/config"
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
	ctx := context.Background()

	config, err := config.Load()
	if err != nil {
		log.Fatalf("[FAILURE] Failed to load the config: %v", err)
	}

	// db connection
	db, err := store.NewStore(ctx, config.PG_DB_URL)
	if err != nil {
		log.Fatalf("[FAILURE] Failed to ping the store: %v", err)
	}
	defer db.Close()
	log.Println("[SUCCESS] Connected to Database")

	// worker pool
	workerPool := worker.NewPool(db, numWorkers, pollInterval)
	workerPool.Start(ctx)
	defer workerPool.Stop()

	// grpc connection
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPC_PORT))
	if err != nil {
		log.Fatalf("[FAILURE] Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterJobSchedulerServer(grpcServer, api.NewServer(db))

	log.Printf("[SUCCESS] gRPC server listening on :%s", config.GRPC_PORT)

	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("[INFO] Shutting down gRPC server...")
	grpcServer.GracefulStop()

	log.Println("[SUCCESS] Server stopped. Bye!")

}
