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
	registry     *Registry
	numWorkers   int
	pollInterval time.Duration
	stopCh       chan struct{}
	jobCh        chan store.Job
	wg           sync.WaitGroup
}

func NewPool(s *store.Store, registry *Registry, numWorkers int, pollInterval time.Duration) *Pool {
	return &Pool{
		store:        s,
		numWorkers:   numWorkers,
		registry:     registry,
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
	logger.Info("Worker processing with payload", "worker", workerId, "job_payload", job.Payload)

	handler, err := p.registry.Get(job.Type)
	if err != nil {
		logger.Error("Worker failed to mark job complete", "worker_id", workerId, "job_id", job.ID, "error", err)
		p.store.UpdateJobStatus(ctx, store.JobStatusFailed, job.ID)
		return
	}

	err = handler.Handle(ctx, job)
	if err != nil {
		logger.Error("Worker failed to mark job complete", "worker_id", workerId, "job_id", job.ID, "error", err)
		p.store.UpdateJobStatus(ctx, store.JobStatusFailed, job.ID)
		return
	}

	p.store.UpdateJobStatus(ctx, store.JobStatusCompleted, job.ID)

	logger.Info("Worker completed the job with payload", "worker", workerId, "job_payload", job.Payload)

}
