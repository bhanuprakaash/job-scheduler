package archive

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type MockUploader struct {
	UploadFunc func(ctx context.Context, data io.Reader, size int64, path, contentType string) error
	ExistsFunc func(ctx context.Context, path string) (bool, error)
}

func (m *MockUploader) Upload(ctx context.Context, data io.Reader, size int64, path, contentType string) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, data, size, path, contentType)
	}
	return nil
}

func (m *MockUploader) Exists(ctx context.Context, path string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, path)
	}
	return false, nil
}

type MockStore struct {
	GetArchivedJobsFunc func(ctx context.Context, duration time.Duration, limit int) ([]store.Job, error)
	BatchDeleteJobsFunc func(ctx context.Context, ids []int64) error
	// Embed the interface to satisfy unused methods automatically (panic if called unexpectedly)
	store.Storer
}

func (m *MockStore) GetArchivedJobs(ctx context.Context, d time.Duration, l int) ([]store.Job, error) {
	if m.GetArchivedJobsFunc != nil {
		return m.GetArchivedJobsFunc(ctx, d, l)
	}
	return nil, nil
}

func (m *MockStore) BatchDeleteJobs(ctx context.Context, ids []int64) error {
	if m.BatchDeleteJobsFunc != nil {
		return m.BatchDeleteJobsFunc(ctx, ids)
	}
	return nil
}

// --- Tests ---

func TestArchiveJob_Handle_Success(t *testing.T) {
	logger.Init()
	// 1. Setup Data
	mockJobs := []store.Job{
		{ID: 101, Status: "completed"},
		{ID: 102, Status: "completed"},
	}

	// 2. Setup Mocks
	uploaded := false
	deleted := false

	storeMock := &MockStore{
		GetArchivedJobsFunc: func(ctx context.Context, d time.Duration, l int) ([]store.Job, error) {
			return mockJobs, nil
		},
		BatchDeleteJobsFunc: func(ctx context.Context, ids []int64) error {
			if len(ids) == 2 && ids[0] == 101 && ids[1] == 102 {
				deleted = true
			}
			return nil
		},
	}

	uploaderMock := &MockUploader{
		UploadFunc: func(ctx context.Context, data io.Reader, size int64, path, contentType string) error {
			uploaded = true
			return nil
		},
	}

	job := NewArchiveJob(storeMock, uploaderMock)

	// 3. Execute
	err := job.Handle(context.Background(), store.Job{
		Payload: `{"older_than": "24h", "batch": 100}`,
	})

	// 4. Verify
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !uploaded {
		t.Error("Expected archive file to be uploaded")
	}
	if !deleted {
		t.Error("Expected processed jobs to be deleted from DB")
	}
}

func TestArchiveJob_Handle_NoJobs(t *testing.T) {
	logger.Init()
	// Setup: Store returns empty list
	storeMock := &MockStore{
		GetArchivedJobsFunc: func(ctx context.Context, d time.Duration, l int) ([]store.Job, error) {
			return []store.Job{}, nil
		},
	}

	uploaderMock := &MockUploader{
		UploadFunc: func(ctx context.Context, data io.Reader, size int64, path, contentType string) error {
			t.Error("Upload should NOT be called when no jobs exist")
			return nil
		},
	}

	job := NewArchiveJob(storeMock, uploaderMock)

	err := job.Handle(context.Background(), store.Job{
		Payload: `{"older_than": "24h", "batch": 100}`,
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestArchiveJob_Handle_UploadFails_NoDelete(t *testing.T) {
	logger.Init()

	storeMock := &MockStore{
		GetArchivedJobsFunc: func(ctx context.Context, d time.Duration, l int) ([]store.Job, error) {
			return []store.Job{{ID: 1}}, nil
		},
		BatchDeleteJobsFunc: func(ctx context.Context, ids []int64) error {
			t.Error("BatchDeleteJobs should NOT be called if upload fails")
			return nil
		},
	}

	uploaderMock := &MockUploader{
		UploadFunc: func(ctx context.Context, data io.Reader, size int64, path, contentType string) error {
			return errors.New("s3 bucket unavailable")
		},
	}

	job := NewArchiveJob(storeMock, uploaderMock)

	err := job.Handle(context.Background(), store.Job{
		Payload: `{"older_than": "24h", "batch": 100}`,
	})

	if err == nil {
		t.Error("Expected error from upload failure, got nil")
	}
}
