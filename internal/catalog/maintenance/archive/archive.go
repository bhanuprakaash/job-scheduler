package maintenance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/blob"
	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type ArchiveJobPayload struct {
	OlderThanStr string `json:"older_than"`
	BatchSize    int    `json:"batch"`
}

type ArchiveJob struct {
	store    store.Storer
	uploader blob.Uploader
}

func NewArchiveJob(store store.Storer, uploader blob.Uploader) *ArchiveJob {
	return &ArchiveJob{
		store:    store,
		uploader: uploader,
	}
}

func (a *ArchiveJob) Handle(ctx context.Context, job store.Job) error {

	var payload ArchiveJobPayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return err
	}
	startTime := time.Now()

	olderThan, err := time.ParseDuration(payload.OlderThanStr)
	if err != nil {
		return fmt.Errorf("invalid duration format '%s': %w", payload.OlderThanStr, err)
	}

	jobs, err := a.store.GetArchivedJobs(ctx, olderThan, payload.BatchSize)
	if err != nil {
		return fmt.Errorf("get archive jobs: %w", err)
	}

	if len(jobs) == 0 {
		logger.Info("No jobs to archive")
		return nil
	}

	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal jobs: %w", err)
	}

	filename := fmt.Sprintf("archives/jobs_%s.json", time.Now().Format("2006-01-02_15-04-05"))

	err = a.uploader.Upload(
		ctx,
		bytes.NewReader(data),
		int64(len(data)),
		filename,
		"application/json",
	)

	if err != nil {
		return fmt.Errorf("failed to upload archive: %w", err)
	}

	var ids []int64
	for _, j := range jobs {
		ids = append(ids, j.ID)
	}

	if err = a.store.BatchDeleteJobs(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete archived jobs: %w", err)
	}

	duration := time.Since(startTime)

	logger.Info("Archived and Deleted jobs", "count", len(ids), "duration", duration)

	return nil

}
