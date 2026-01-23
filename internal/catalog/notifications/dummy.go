package notifications

import (
	"context"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type DummyJob struct{}

func (d *DummyJob) Handle(ctx context.Context, job store.Job) error {
	logger.Info("Processing Job", "job", job.ID)
	time.Sleep(3 * time.Second)
	return nil
}
