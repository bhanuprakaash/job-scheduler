package email

import (
	"context"
	"log/slog"
	"time"
)

type SendConsoleEmail struct {
	logger *slog.Logger
}

func NewSendConsoleEmail(logger *slog.Logger) *SendConsoleEmail {
	return &SendConsoleEmail{
		logger: logger,
	}
}

func (c *SendConsoleEmail) Send(ctx context.Context, to, subject, body string) error {
	c.logger.Info("sending the email", "to", to, "subject", subject)
	time.Sleep(3 * time.Second)
	return nil
}
