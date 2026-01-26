package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bhanuprakaash/job-scheduler/internal/config"
	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	pb "github.com/bhanuprakaash/job-scheduler/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func runHTTPServer(ctx context.Context, cfg *config.Config) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	grpcEndpoint := fmt.Sprintf("localhost:%s", cfg.GRPC_PORT)

	err := pb.RegisterJobSchedulerHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		logger.Fatal("Faild to register gateway", "error", err)
	}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"}, // Allow your Vite frontend
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	handler := c.Handler(mux)

	httpPort := cfg.HTTP_PORT
	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: handler,
	}

	logger.Info("HTTP Gateway listening", "port", httpPort)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to server HTTP", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down HTTP gateway...")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error("Failed to shutdown HTTP server", "error", err)
	}

}
