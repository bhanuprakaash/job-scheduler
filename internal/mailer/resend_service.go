package mailer

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

type ResendEmailService struct {
	client *resend.Client
	from   string
}

func NewResendEmailService(apikey string, from string) *ResendEmailService {
	client := resend.NewClient(apikey)
	return &ResendEmailService{
		client: client,
		from:   from,
	}
}

func (r *ResendEmailService) Send(ctx context.Context, to, subject, body string) error {
	params := &resend.SendEmailRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	_, err := r.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("sending to email: %w", err)
	}

	return nil
}
