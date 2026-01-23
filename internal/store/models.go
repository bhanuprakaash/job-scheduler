package store

import (
	"database/sql"
	"time"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type Job struct {
	ID           int64          `db:"id"`
	Type         string         `db:"type"`
	Payload      string         `db:"payload"`
	Status       JobStatus      `db:"status"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	StartedAt    *time.Time     `db:"started_at"`
	CompletedAt  *time.Time     `db:"completed_at"`
	ErrorMessage sql.NullString `db:"last_err"`
	RetryCount   int            `db:"retry_count"`
}
