package worker

import (
	"context"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type Handler interface {
	Handle(ctx context.Context, job store.Job) error
}

