package worker

import (
	"context"
	"sync"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type Pool struct {
	store        *store.Store
	numWorkers   int
	pollInterval time.Duration
	stopCh       chan struct{}
	jobCh        chan store.Job
	wg           sync.WaitGroup
}

func NewPool(s *store.Store, numWorkers int, pollInterval time.Duration) *Pool {
	return &Pool{
		store:        s,
		numWorkers:   numWorkers,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
		jobCh:        make(chan store.Job, 10),
	}
}

func (p *Pool) Start(ctx context.Context) {
	logger.Info("worker pool started", "workers", p.numWorkers)
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i+1)
	}

	go p.StartDispatcher(ctx)
}

func (p *Pool) Stop() {
	logger.Info("worker pool shutting down")
	close(p.stopCh)
	p.wg.Wait()
	logger.Info("worker pool stopped")

}

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	logger.Info("Worker started: ", "worker", id)

	for job := range p.jobCh {
		logger.Info("Worker picked up job", "id", id, "job_id", job.ID)
		p.ProcessNextJob(ctx, id, job)
	}

	logger.Info("Worker stopping", "id", id)

}

func (p *Pool) StartDispatcher(ctx context.Context) {
	logger.Info("starting dispatcher")
	defer close(p.jobCh)

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			logger.Info("Dispatcher shutting down")
			return

		case <-ticker.C:
			logger.Debug("Ticker fired. Checking DB...")
			jobs, err := p.store.GetPendingJobs(ctx, 10)
			if err != nil {
				logger.Error("fetching jobs", "err", err)
				continue
			}

			if len(jobs) > 0 {
				logger.Info("Dispatcher found jobs", "jobs", len(jobs))
			}

			for _, job := range jobs {
				select {
				case <-p.stopCh:
					return
				case p.jobCh <- job:
					// inserted
				}
			}
		}
	}

}

func (p *Pool) ProcessNextJob(ctx context.Context, workerId int, job store.Job) {
	err := p.store.UpdateJobStatus(ctx, store.JobStatusRunning, job.ID)
	if err != nil {
		logger.Error("Worker failed to mark job", "worker_id", workerId, "job_id", job.ID, "error", err)
		return
	}

	logger.Info("Worker processing with payload", "worker", workerId, "job_payload", job.Payload)
	time.Sleep(2 * time.Second)

	err = p.store.UpdateJobStatus(ctx, store.JobStatusCompleted, job.ID)
	if err != nil {
		logger.Error("Worker failed to mark job complete", "worker_id", workerId, "job_id", job.ID, "error", err)
		return
	}
}
