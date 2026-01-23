package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/config"
	pb "github.com/bhanuprakaash/job-scheduler/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var serverAddr string

func main() {
	rootCmd := &cobra.Command{
		Use:   "job-cli",
		Short: "Job Scheduler CLI",
	}
	config, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load the config: %v", err)
	}
	grpcHost := fmt.Sprintf("localhost:%s", config.GRPC_PORT)

	rootCmd.PersistentFlags().StringVar(&serverAddr, "server", grpcHost, "gRPC server address")

	rootCmd.AddCommand(submitCmd())
	rootCmd.AddCommand(getCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func submitCmd() *cobra.Command {
	var jobType, payload string

	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a new job",
		Run: func(cmd *cobra.Command, args []string) {
			conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("Failed to connect: %v", err)
			}
			defer conn.Close()
			client := pb.NewJobSchedulerClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.SubmitJob(ctx, &pb.SubmitJobRequest{
				Type:    jobType,
				Payload: payload,
			})
			if err != nil {
				log.Fatalf("Failed to submit job: %v", err)
			}

			fmt.Printf("✓ Job submitted successfully\n")
			fmt.Printf("  Job ID: %s\n", resp.JobId)
			fmt.Printf("  Status: %s\n", resp.Status)
		},
	}

	cmd.Flags().StringVar(&jobType, "type", "dummy", "Job type")
	cmd.Flags().StringVar(&payload, "data", "{}", "Job payload (JSON)")

	return cmd
}

func getCmd() *cobra.Command {
	var jobID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get job status",
		Run: func(cmd *cobra.Command, args []string) {
			conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("Failed to connect: %v", err)
			}
			defer conn.Close()

			client := pb.NewJobSchedulerClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.GetJob(ctx, &pb.GetJobRequest{JobId: jobID})
			if err != nil {
				log.Fatalf("Failed to get job: %v", err)
			}

			fmt.Printf("Job Details:\n")
			fmt.Printf("  ID:         %s\n", resp.JobId)
			fmt.Printf("  Type:       %s\n", resp.Type)
			fmt.Printf("  Status:     %s\n", resp.Status)
			fmt.Printf("  Payload:    %s\n", resp.Payload)
			fmt.Printf("  Created:    %s\n", resp.CreatedAt)
			if resp.CompletedAt != "" {
				fmt.Printf("  Completed:  %s\n", resp.CompletedAt)
			}
			if resp.Status == "failed" {
				fmt.Printf("  Error Message:  %s\n",resp.ErrorMessage)
			}

		},
	}

	cmd.Flags().StringVar(&jobID, "id", "", "Job ID (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}
