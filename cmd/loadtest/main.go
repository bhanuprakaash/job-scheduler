package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/bhanuprakaash/job-scheduler/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Address     string
	TotalJobs   int
	Concurrency int
}

func main() {
	// 1. Config
	addr := flag.String("addr", "localhost:50052", "gRPC server address")
	total := flag.Int("total", 1000, "Total number of jobs to submit")
	workers := flag.Int("workers", 50, "Number of concurrent workers")
	flag.Parse()

	cfg := Config{
		Address:     *addr,
		TotalJobs:   *total,
		Concurrency: *workers,
	}

	fmt.Printf("🚀 Starting Load Test (Invoice Gen): %d jobs, %d workers\n", cfg.TotalJobs, cfg.Concurrency)

	// 2. Connect
	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewJobSchedulerClient(conn)

	// 3. Stats
	var (
		wg           sync.WaitGroup
		jobsSent     atomic.Int64
		successCount atomic.Int64
		failCount    atomic.Int64
		durations    = make([]time.Duration, 0, cfg.TotalJobs)
		durationsMu  sync.Mutex
		jobChan      = make(chan int, cfg.TotalJobs)
	)

	startTime := time.Now()

	// 4. Start Workers
	wg.Add(cfg.Concurrency)
	for i := 0; i < cfg.Concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			for jobID := range jobChan {
				start := time.Now()

				// --- PAYLOAD: Finance Invoice ---
				// Note: invoice_id MUST be unique to bypass idempotency checks and force PDF gen.
				req := &pb.SubmitJobRequest{
					Type: "finance:invoice",
					Payload: fmt.Sprintf(`{
						"user_id": "load_user_%d",
						"amount": 99.99,
						"currency": "USD",
						"date": "2026-01-28",
						"invoice_id": "inv_load_%d", 
						"items": [
							{ "description": "Load Test Item A", "quantity": 1, "unit_price": 50.00 },
							{ "description": "Load Test Item B", "quantity": 2, "unit_price": 24.995 }
						]
					}`, workerID, jobID),
				}

				// RPC Call
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				_, err := client.SubmitJob(ctx, req)
				cancel()

				latency := time.Since(start)

				if err != nil {
					failCount.Add(1)
					// log.Printf("Error: %v", err) // Uncomment to debug
				} else {
					successCount.Add(1)
				}

				durationsMu.Lock()
				durations = append(durations, latency)
				durationsMu.Unlock()

				current := jobsSent.Add(1)
				if current%100 == 0 {
					fmt.Printf("\rProgress: %d/%d jobs...", current, cfg.TotalJobs)
				}
			}
		}(i)
	}

	// 5. Fill Queue
	for i := 0; i < cfg.TotalJobs; i++ {
		jobChan <- i
	}
	close(jobChan)

	wg.Wait()
	totalTime := time.Since(startTime)
	fmt.Println("\n\n✅ Load Test Completed!")

	// 6. Report
	durationsMu.Lock()
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	var p50, p99 time.Duration
	if len(durations) > 0 {
		p50 = durations[len(durations)/2]
		p99 = durations[int(float64(len(durations))*0.99)]
	}
	durationsMu.Unlock()

	fmt.Println("=========================================")
	fmt.Printf("Total Duration:   %v\n", totalTime)
	fmt.Printf("Throughput:       %.2f jobs/sec\n", float64(cfg.TotalJobs)/totalTime.Seconds())
	fmt.Printf("Success:          %d\n", successCount.Load())
	fmt.Printf("Failed:           %d\n", failCount.Load())
	fmt.Printf("Latency P50:      %v\n", p50)
	fmt.Printf("Latency P99:      %v\n", p99)
	fmt.Println("=========================================")
}
