package email

import (
	"context"
	"errors"
	"testing"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type MockSender struct {
	SendFunc func(ctx context.Context, to, subject, body string) error
}

func (m *MockSender) Send(ctx context.Context, to, subject, body string) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, to, subject, body)
	}
	return nil
}

func TestEmailJob_Handle_Success(t *testing.T) {
	logger.Init()
	emailSent := false
	mockSender := &MockSender{
		SendFunc: func(ctx context.Context, to, subject, body string) error {
			if to == "test@example.com" && subject == "Hello" {
				emailSent = true
			}
			return nil
		},
	}

	jobHandler := NewEmailJob(mockSender)

	err := jobHandler.Handle(context.Background(), store.Job{
		ID:      1,
		Payload: `{"to": "test@example.com", "subject": "Hello", "body": "World"}`,
	})

	if err != nil {
		t.Errorf("Handle() returned error: %v", err)
	}
	if !emailSent {
		t.Error("Email was not sent to the correct recipient")
	}
}

func TestEmailJob_Handle_Failure_Retries(t *testing.T) {
	logger.Init()
	mockSender := &MockSender{
		SendFunc: func(ctx context.Context, to, subject, body string) error {
			return errors.New("API outage")
		},
	}

	jobHandler := NewEmailJob(mockSender)

	err := jobHandler.Handle(context.Background(), store.Job{
		ID:      1,
		Payload: `{"to": "test@example.com"}`,
	})

	if err == nil {
		t.Error("Expected error when email service fails, got nil")
	}
}
