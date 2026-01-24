package email

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/mailer"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
type EmailJob struct {
	sender mailer.Sender
}

func NewEmailJob(sender mailer.Sender) *EmailJob {
	return &EmailJob{
		sender: sender,
	}
}

func (e *EmailJob) Handle(ctx context.Context, job store.Job) error {
	var payload EmailPayload
	startTime := time.Now()
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return fmt.Errorf("failed to parse email payload: %w", err)
	}

	logger.Info("Sending email",
		"job_id", job.ID,
		"to", payload.To,
		"subject", payload.Subject,
		"retry", job.RetryCount)

	err := e.sender.Send(ctx, payload.To, payload.Subject, payload.Body)
	if err != nil {
		logger.Error("Email send failed",
			"job_id", job.ID,
			"to", payload.To,
			"duration", time.Since(startTime),
			"error", err)
		return fmt.Errorf("error sending mail: %w", err)
	}
	logger.Info("Email sent successfully",
		"job_id", job.ID,
		"to", payload.To,
		"duration", time.Since(startTime))
	return nil
}
