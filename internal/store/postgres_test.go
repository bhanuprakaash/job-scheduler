package store

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
)


func setupIntegrationTest(t *testing.T) *Store {
	logger.Init()
	// dbUrl := os.Getenv("PG_DB_URL")
	dbUrl := "postgres://postgres:postgres@localhost:5433/job_scheduler?sslmode=disable"
	if dbUrl == "" {
		t.Skip("Skipping integration test: PG_DB_URL environment variable not set")
	}

	ctx := context.Background()
	store, err := NewStore(ctx, dbUrl)
	if err != nil {
		t.Fatalf("Failed to connect to integration DB: %v", err)
	}

	// 🧹 Cleanup: Truncate table to ensure a clean state
	// RESTART IDENTITY resets the ID counter to 1
	_, err = store.db.Exec(ctx, "TRUNCATE TABLE jobs RESTART IDENTITY")
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	return store
}

func TestIntegration_CreateJob(t *testing.T) {
	s := setupIntegrationTest(t)
	defer s.Close()
	ctx := context.Background()

	// 1. Create a Job
	job, err := s.CreateJob(ctx, "test:create", `{"foo": "bar"}`)
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}

	// 2. Verify fields
	if job.ID == 0 {
		t.Error("Expected Job ID to be generated, got 0")
	}
	if job.Status != JobStatusPending {
		t.Errorf("Expected status 'pending', got '%s'", job.Status)
	}
	if job.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestIntegration_GetPendingJobs_Concurrency(t *testing.T) {
	// 🧪 THE CRITICAL TEST: Validating "SKIP LOCKED"
	// This proves that multiple workers won't steal each other's jobs.
	s := setupIntegrationTest(t)
	defer s.Close()
	ctx := context.Background()

	// 1. Seed 20 jobs
	totalJobs := 20
	for i := 0; i < totalJobs; i++ {
		_, err := s.CreateJob(ctx, "test:concurrent", fmt.Sprintf(`{"index": %d}`, i))
		if err != nil {
			t.Fatal(err)
		}
	}

	// 2. Launch 4 concurrent "Worker" Goroutines
	// Each worker tries to fetch 5 jobs.
	numWorkers := 4
	jobsPerWorker := 5

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Channel to collect results from all workers
	resultsCh := make(chan []Job, numWorkers)
	errorsCh := make(chan error, numWorkers)

	for w := 0; w < numWorkers; w++ {
		go func(workerID int) {
			defer wg.Done()

			// Simulate the Worker Loop logic
			jobs, err := s.GetPendingJobs(ctx, jobsPerWorker)
			if err != nil {
				errorsCh <- fmt.Errorf("worker %d failed: %w", workerID, err)
				return
			}
			resultsCh <- jobs
		}(w)
	}

	wg.Wait()
	close(resultsCh)
	close(errorsCh)

	// 3. Check for Errors
	for err := range errorsCh {
		t.Fatalf("Concurrency error: %v", err)
	}

	// 4. Validate Results (Mutual Exclusion)
	seenJobs := make(map[int64]bool)
	totalFetched := 0

	for workerJobs := range resultsCh {
		if len(workerJobs) != jobsPerWorker {
			t.Errorf("Expected worker to get %d jobs, got %d", jobsPerWorker, len(workerJobs))
		}

		for _, job := range workerJobs {
			// CRITICAL CHECK: Has this job been seen before?
			if seenJobs[job.ID] {
				t.Errorf("🚨 DATA RACE: Job %d was processed by multiple workers!", job.ID)
			}
			seenJobs[job.ID] = true
			totalFetched++

			// Ensure state transition happened
			if job.Status != JobStatusRunning {
				t.Errorf("Job %d should be 'running', got '%s'", job.ID, job.Status)
			}
		}
	}

	if totalFetched != totalJobs {
		t.Errorf("Expected to fetch %d total jobs, got %d", totalJobs, totalFetched)
	}
}

func TestIntegration_HandleJobFailure_RetryLogic(t *testing.T) {
	s := setupIntegrationTest(t)
	defer s.Close()
	ctx := context.Background()

	// 1. Create a job
	job, _ := s.CreateJob(ctx, "test:retry", "{}")

	// 2. Move to 'running' manually (simulating a pickup)
	_, err := s.db.Exec(ctx, "UPDATE jobs SET status = 'running' WHERE id = $1", job.ID)
	if err != nil {
		t.Fatal(err)
	}

	// 3. Fail it (Retry 1)
	errMsg := "network timeout"
	err = s.HandleJobFailure(ctx, job.ID, errMsg)
	if err != nil {
		t.Fatalf("HandleJobFailure failed: %v", err)
	}

	// 4. Verify Job State (Should be Pending + Backoff)
	updatedJob, err := s.GetJobByID(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}

	if updatedJob.Status != JobStatusPending {
		t.Errorf("Expected status 'pending' (for retry), got '%s'", updatedJob.Status)
	}
	if updatedJob.RetryCount != 1 {
		t.Errorf("Expected retry_count 1, got %d", updatedJob.RetryCount)
	}
	if updatedJob.ErrorMessage.String != errMsg {
		t.Errorf("Expected error message saved, got %s", updatedJob.ErrorMessage.String)
	}

	// 5. Verify Backoff (Next Run should be in the future)
	// We have to query raw SQL because GetJobByID doesn't return next_run_at
	var nextRunAt time.Time
	err = s.db.QueryRow(ctx, "SELECT next_run_at FROM jobs WHERE id=$1", job.ID).Scan(&nextRunAt)
	if err != nil {
		t.Fatal(err)
	}

	if nextRunAt.Before(time.Now()) {
		t.Error("Backoff failed: next_run_at should be in the future")
	}
}

func TestIntegration_HandleJobFailure_MaxRetries(t *testing.T) {
	s := setupIntegrationTest(t)
	defer s.Close()
	ctx := context.Background()

	// 1. Create a job and force retry_count to 2 (Assuming Max=3)
	job, _ := s.CreateJob(ctx, "test:max_retry", "{}")
	_, err := s.db.Exec(ctx, "UPDATE jobs SET status = 'running', retry_count = 2 WHERE id = $1", job.ID)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Fail it (This should be the 3rd strike)
	err = s.HandleJobFailure(ctx, job.ID, "fatal error")
	if err != nil {
		t.Fatal(err)
	}

	// 3. Verify it is DEAD
	deadJob, _ := s.GetJobByID(ctx, job.ID)

	if deadJob.Status != JobStatusFailed {
		t.Errorf("Expected status 'failed', got '%s'", deadJob.Status)
	}
	if deadJob.RetryCount != 3 {
		t.Errorf("Expected retry_count 3, got %d", deadJob.RetryCount)
	}
}
