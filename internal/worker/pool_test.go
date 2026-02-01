package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type MemoryStore struct {
	mu       sync.Mutex
	jobs     []store.Job
	finished map[int64]store.JobStatus
}

func NewMemoryStore(jobs []store.Job) *MemoryStore {
	return &MemoryStore{
		jobs:     jobs,
		finished: make(map[int64]store.JobStatus),
	}
}

func (m *MemoryStore) GetPendingJobs(ctx context.Context, limit int) ([]store.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.jobs) == 0 {
		return nil, nil
	}

	count := limit
	if count > len(m.jobs) {
		count = len(m.jobs)
	}

	batch := m.jobs[:count]
	m.jobs = m.jobs[count:]

	return batch, nil
}

func (m *MemoryStore) UpdateJobStatus(ctx context.Context, status store.JobStatus, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.finished[id] = status
	return nil
}

func (m *MemoryStore) CreateJob(ctx context.Context, jobType, payload string) (*store.Job, error) {
	return nil, nil
}
func (m *MemoryStore) GetJobByID(ctx context.Context, id int64) (*store.Job, error) { return nil, nil }
func (m *MemoryStore) GetArchivedJobs(ctx context.Context, d time.Duration, l int) ([]store.Job, error) {
	return nil, nil
}
func (m *MemoryStore) HandleJobFailure(ctx context.Context, id int64, errMsg string) error {
	return nil
}
func (m *MemoryStore) BatchDeleteJobs(ctx context.Context, ids []int64) error { return nil }
func (m *MemoryStore) ListJobs(ctx context.Context, limit, offset int) (*store.PaginatedJobs, error) {
	return nil, nil
}
func (m *MemoryStore) GetStats(ctx context.Context) (*store.JobStats, error) { return nil, nil }
func (m *MemoryStore) Close()                                                {}
func (m *MemoryStore) RepeatStuckJobs(ctx context.Context, interval time.Duration) (int64, error) {
	return 0, nil
}

type HandlerFunc func(ctx context.Context, job store.Job) error

func (f HandlerFunc) Handle(ctx context.Context, job store.Job) error {
	return f(ctx, job)
}

func TestPool_LoadTest(t *testing.T) {
	logger.Init()

	const (
		TotalJobs   = 100
		WorkerCount = 10
		PollTime    = 5 * time.Millisecond
	)

	var initialJobs []store.Job
	for i := 1; i <= TotalJobs; i++ {
		initialJobs = append(initialJobs, store.Job{
			ID:      int64(i),
			Type:    "load:test",
			Payload: "{}",
			Status:  store.JobStatusPending,
		})
	}

	memStore := NewMemoryStore(initialJobs)
	registry := NewRegistry()

	var wg sync.WaitGroup
	wg.Add(TotalJobs)

	registry.Register("load:test", HandlerFunc(func(ctx context.Context, j store.Job) error {
		defer wg.Done()
		time.Sleep(2 * time.Millisecond)
		return nil
	}), 0)

	pool := NewPool(memStore, registry, WorkerCount, PollTime)
	ctx, cancel := context.WithCancel(context.Background())

	startTime := time.Now()
	pool.Start(ctx)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		elapsed := time.Since(startTime)
		t.Logf("✓ Processed %d jobs with %d workers in %v", TotalJobs, WorkerCount, elapsed)
	case <-time.After(5 * time.Second):
		t.Fatal("✗ Timeout: Not all jobs were processed")
	}

	cancel()
	pool.Stop()

	memStore.mu.Lock()
	defer memStore.mu.Unlock()
	if len(memStore.finished) != TotalJobs {
		t.Errorf("Expected %d completed jobs, got %d", TotalJobs, len(memStore.finished))
	}
}

func TestPool_GracefulShutdown(t *testing.T) {
	logger.Init()

	jobs := []store.Job{{ID: 1, Type: "long:job"}}
	memStore := NewMemoryStore(jobs)

	registry := NewRegistry()

	started := make(chan struct{})
	registry.Register("long:job", HandlerFunc(func(ctx context.Context, j store.Job) error {
		close(started)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			return nil
		}
	}), 0)

	pool := NewPool(memStore, registry, 1, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	pool.Start(ctx)

	<-started

	cancel()

	pool.Stop()

	t.Log("✓ Pool shut down gracefully")
}
