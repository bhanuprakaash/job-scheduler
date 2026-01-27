package worker

import (
	"context"
	"sync"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/metrics"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type Pool struct {
	store        store.Storer
	registry     *Registry
	numWorkers   int
	pollInterval time.Duration
	stopCh       chan struct{}
	jobCh        chan store.Job
	wg           sync.WaitGroup
}

func NewPool(s store.Storer, registry *Registry, numWorkers int, pollInterval time.Duration) *Pool {
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
	logger.Info("Worker processing the", "worker", workerId, "job_id", job.ID)

	metrics.ActiveWorkers.Inc()
	defer metrics.ActiveWorkers.Dec()

	startTime := time.Now()

	handler, err := p.registry.Get(job.Type)
	if err != nil {
		logger.Error("no handler found", "error", err)
		updateFail := p.store.UpdateJobStatus(ctx, store.JobStatusFailed, job.ID)
		if updateFail != nil {
			logger.Error("CRITICAL: Failed to update job status", "error", updateFail)
		}
		metrics.JobsProcessed.WithLabelValues(job.Type, "failed").Inc()
		return
	}

	err = handler.Handle(ctx, job)
	duration := time.Since(startTime).Seconds()
	metrics.JobDuration.WithLabelValues(job.Type).Observe(duration)
	if err != nil {
		logger.Error("Job failed ", "worker_id", workerId, "job_id", job.ID, "error", err)
		failErr := p.store.HandleJobFailure(ctx, job.ID, err.Error())
		if failErr != nil {
			logger.Error("CRITICAL: Failed to update job status", "error", failErr)
		}
		metrics.JobsProcessed.WithLabelValues(job.Type, "failed").Inc()
		return
	}

	updateFail := p.store.UpdateJobStatus(ctx, store.JobStatusCompleted, job.ID)
	if updateFail != nil {
		logger.Error("CRITICAL: Failed to update job status", "error", updateFail)
		return
	}

	metrics.JobsProcessed.WithLabelValues(job.Type, "success").Inc()

	logger.Info("Worker completed the job", "worker", workerId, "job_id", job.ID)

}
