package resize

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"time"

	_ "image/gif"
	_ "image/png"

	"github.com/bhanuprakaash/job-scheduler/internal/blob"
	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
	"golang.org/x/image/draw"
)

type imageResizePayload struct {
	ImageSrc   string `json:"src_url"`
	Width      int    `json:"width"`
	OutputPath string `json:"output_path"`
}

type ImageResizeJob struct {
	blobUploader blob.BlobUploader
}

func NewImageResizeJob(blobUploader blob.BlobUploader) *ImageResizeJob {
	return &ImageResizeJob{
		blobUploader: blobUploader,
	}
}

func (r *ImageResizeJob) Handle(ctx context.Context, job store.Job) error {
	var payload imageResizePayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	start := time.Now()

	resp, err := http.Get(payload.ImageSrc)
	if err != nil {
		return fmt.Errorf("error downloading image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("received non-200 status code:%w", err)
	}

	srcImg, _, err := image.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := srcImg.Bounds()
	ratio := float64(bounds.Dy()) / float64(bounds.Dx())
	newHeight := int(float64(payload.Width) * ratio)

	dstImg := image.NewRGBA(image.Rect(0, 0, payload.Width, newHeight))
	draw.Draw(dstImg, dstImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.CatmullRom.Scale(dstImg, dstImg.Bounds(), srcImg, bounds, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dstImg, &jpeg.Options{Quality: 80}); err != nil {
		return fmt.Errorf("failed to encode jpeg: %w", err)
	}

	err = r.blobUploader.Upload(ctx, &buf, int64(buf.Len()), payload.OutputPath, "image/jpeg")

	duration := time.Since(start)

	logger.Info("Image Resized Successfully", "job_id", job.ID, "duration", duration, "output path", payload.OutputPath)

	return nil

}
