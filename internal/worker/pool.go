package worker

import (
	"context"
	"log"
	"sync"
	"time"

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
	log.Printf("[INFO] worker pool started with %d workers", p.numWorkers)
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i+1)
	}

	go p.StartDispatcher(ctx)
}

func (p *Pool) Stop() {
	log.Println("[INFO] worker pool shutting down....")
	close(p.stopCh)
	p.wg.Wait()
	log.Println("[INFO] worker pool stopped")

}

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	log.Printf("[INFO] Worker-%d started: ", id)

	for job := range p.jobCh {
		log.Printf("[INFO] Worker-%d picked up job %d", id, job.ID)
		p.ProcessNextJob(ctx, id, job)
	}

	log.Printf("[INFO] Worker-%d stopping", id)

}

func (p *Pool) StartDispatcher(ctx context.Context) {
	log.Println("[INFO] starting dispatcher...")
	defer close(p.jobCh)

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			log.Printf("[INFO] Dispatcher shutting down")
			return

		case <-ticker.C:
			log.Println("[DEBUG] Ticker fired. Checking DB...")
			jobs, err := p.store.GetPendingJobs(ctx, 10)
			if err != nil {
				log.Printf("[ERROR] fetching jobs: %v", err)
				continue
			}

			if len(jobs) > 0 {
				log.Printf("[DISPATCHER] found %d jobs", len(jobs))
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
		log.Printf("[ERROR] Worker-%d failed to mark job %d running: %v", workerId, job.ID, err)
		return
	}

	log.Printf("[WORKER] Worker-%d processing payload: %s", workerId, job.Payload)
	time.Sleep(2 * time.Second)

	err = p.store.UpdateJobStatus(ctx, store.JobStatusCompleted, job.ID)
	if err != nil {
		log.Printf("[ERROR] Worker-%d failed to mark job %d completed: %v", workerId, job.ID, err)
		return
	}
}
