package store

import (
	"context"
	"time"
)

type Storer interface {
	CreateJob(ctx context.Context, jobType, payload string) (*Job, error)
	GetJobByID(ctx context.Context, id int64) (*Job, error)
	GetPendingJobs(ctx context.Context, limit int) ([]Job, error)
	UpdateJobStatus(ctx context.Context, status JobStatus, id int64) error
	HandleJobFailure(ctx context.Context, jobId int64, errMsg string) error
	GetArchivedJobs(ctx context.Context, duration time.Duration, limit int) ([]Job, error)
	BatchDeleteJobs(ctx context.Context, ids []int64) error
	Close()
}
