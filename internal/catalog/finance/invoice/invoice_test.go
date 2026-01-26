package invoice

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

// Reuse the MockUploader pattern (or define a specific StorageClient mock)
type MockStorageClient struct {
	UploadFunc func(ctx context.Context, data io.Reader, size int64, path, contentType string) error
	ExistsFunc func(ctx context.Context, path string) (bool, error)
}

func (m *MockStorageClient) Upload(ctx context.Context, data io.Reader, size int64, path, contentType string) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, data, size, path, contentType)
	}
	return nil
}

func (m *MockStorageClient) Exists(ctx context.Context, path string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, path)
	}
	return false, nil
}

func TestInvoiceJob_Handle_GeneratesPDF(t *testing.T) {
	logger.Init()
	// 1. Setup
	uploaded := false
	mockClient := &MockStorageClient{
		ExistsFunc: func(ctx context.Context, path string) (bool, error) {
			return false, nil // Invoice does not exist yet
		},
		UploadFunc: func(ctx context.Context, data io.Reader, size int64, path, contentType string) error {
			if !strings.HasPrefix(path, "secure/invoices/") {
				t.Errorf("Invalid path: %s", path)
			}
			if contentType != "application/pdf" {
				t.Errorf("Invalid content type: %s", contentType)
			}
			if size == 0 {
				t.Error("PDF file size is 0")
			}
			uploaded = true
			return nil
		},
	}

	job := NewInvoiceJob(mockClient)

	// 2. Execute
	payload := `{"user_id": "u1", "invoice_id": "inv_001", "date": "2026-01-27", "amount": 100.00, "currency": "USD", "items": [{"description": "Service", "quantity": 1, "unit_price": 100}]}`
	err := job.Handle(context.Background(), store.Job{
		Payload: payload,
	})

	// 3. Verify
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !uploaded {
		t.Error("Expected PDF to be uploaded")
	}
}

func TestInvoiceJob_Handle_Idempotency(t *testing.T) {
	logger.Init()
	// 1. Setup: Simulate Invoice ALREADY Exists
	mockClient := &MockStorageClient{
		ExistsFunc: func(ctx context.Context, path string) (bool, error) {
			return true, nil 
		},
		UploadFunc: func(ctx context.Context, data io.Reader, size int64, path, contentType string) error {
			t.Error("Upload should NOT be called if invoice exists")
			return nil
		},
	}

	job := NewInvoiceJob(mockClient)

	// 2. Execute
	err := job.Handle(context.Background(), store.Job{
		Payload: `{"invoice_id": "inv_existing"}`,
	})

	// 3. Verify
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// The assertion is inside the mock (UploadFunc fails test if called)
}