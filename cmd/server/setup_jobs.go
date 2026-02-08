package main

import (
	"github.com/bhanuprakaash/job-scheduler/internal/blob"
	"github.com/bhanuprakaash/job-scheduler/internal/catalog/finance/invoice"
	maintenance "github.com/bhanuprakaash/job-scheduler/internal/catalog/maintenance/archive"
	"github.com/bhanuprakaash/job-scheduler/internal/catalog/media/resize"
	"github.com/bhanuprakaash/job-scheduler/internal/catalog/notifications/email"
	"github.com/bhanuprakaash/job-scheduler/internal/config"
	"github.com/bhanuprakaash/job-scheduler/internal/mailer"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
	"github.com/bhanuprakaash/job-scheduler/internal/worker"
)

func setupJobRegistry(cfg *config.Config, db store.Storer) (*worker.Registry, error) {

	resendService := mailer.NewResendEmailService(cfg.RESEND_EMAIL_API_KEY, cfg.RESEND_FROM_EMAIL)
	minioBlob, err := blob.NewMinioBlob(cfg.MINIO_ID, cfg.MINIO_SECRET, cfg.MINIO_ENDPOINT, cfg.MINIO_BUCKET, cfg.MINIO_USE_SSL)
	if err != nil {
		return nil, err
	}
	jobRegistry := worker.NewRegistry()
	jobRegistry.Register("notification:email", email.NewEmailJob(resendService), 5)
	jobRegistry.Register("media:resize_image", resize.NewImageResizeJob(minioBlob), 2)
	jobRegistry.Register("maintenance:archive", maintenance.NewArchiveJob(db, minioBlob), 0)
	jobRegistry.Register("finance:invoice", invoice.NewInvoiceJob(minioBlob), 10)

	return jobRegistry, nil

}
