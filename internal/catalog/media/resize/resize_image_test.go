package resize

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

// MockUploader simulates MinIO
type MockUploader struct {
	UploadFunc func(ctx context.Context, data io.Reader, size int64, path string, contentType string) error
}

func (m *MockUploader) Upload(ctx context.Context, data io.Reader, size int64, path string, contentType string) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, data, size, path, contentType)
	}
	return nil
}

func (m *MockUploader) Exists(ctx context.Context, path string) (bool, error) { return false, nil }

func TestResizeJob_Handle_Success(t *testing.T) {
	logger.Init()
	// 1. Setup Fake Image Server (Simulating the Internet)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a 100x100 red square on the fly
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				img.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}
		png.Encode(w, img)
	}))
	defer ts.Close()

	// 2. Setup Mock Uploader
	uploaded := false
	mockUploader := &MockUploader{
		UploadFunc: func(ctx context.Context, data io.Reader, size int64, path string, contentType string) error {
			uploaded = true
			if contentType != "image/jpeg" {
				t.Errorf("Expected image/jpeg, got %s", contentType)
			}
			return nil
		},
	}

	jobHandler := NewImageResizeJob(mockUploader)

	// 3. Execution (Point to local test server)
	payload := `{"src_url": "` + ts.URL + `", "width": 50, "output_path": "thumbs/test.jpg"}`
	err := jobHandler.Handle(context.Background(), store.Job{
		ID:      1,
		Payload: payload,
	})

	// 4. Assertion
	if err != nil {
		t.Errorf("Handle() returned error: %v", err)
	}
	if !uploaded {
		t.Error("Resize job did not attempt to upload the result")
	}
}