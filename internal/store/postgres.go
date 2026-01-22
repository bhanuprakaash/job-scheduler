package store

import (
	"context"
	"fmt"
	"time"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(ctx context.Context, databaseUrl string) (*Store, error) {
	config, err := pgxpool.ParseConfig(databaseUrl)

	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Store{db: pool}, nil
}

func (s *Store) Close() {
	s.db.Close()
	logger.Info("db disconnected")
}

func (s *Store) CreateJob(ctx context.Context, jobType, payload string) (*Job, error) {
	var job = &Job{}
	query :=
		`
		INSERT INTO jobs (type, payload)
		VALUES ($1, $2)
		RETURNING id, type, payload, status, created_at, updated_at
		`

	err := s.db.QueryRow(ctx, query, jobType, payload).
		Scan(&job.ID, &job.Type, &job.Payload, &job.Status, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("insert job: %w", err)
	}
	return job, nil
}

func (s *Store) GetJobByID(ctx context.Context, id int64) (*Job, error) {
	var job = &Job{}

	query :=
		`
		SELECT id, type, payload, status, created_at, updated_at, started_at, completed_at
		FROM jobs
		WHERE id = $1
		`
	err := s.db.QueryRow(ctx, query, id).
		Scan(
			&job.ID,
			&job.Type,
			&job.Payload,
			&job.Status,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.StartedAt,
			&job.CompletedAt,
		)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no jobs found with the id: %d", id)
		}
		return nil, fmt.Errorf("get job: %w", err)
	}

	return job, nil
}

func (s *Store) GetPendingJobs(ctx context.Context, limit int) ([]Job, error) {

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query :=
		`
		SELECT id, type, payload, status, created_at, updated_at, started_at, completed_at
		FROM jobs
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
		`
	rows, err := tx.Query(ctx, query, JobStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		err := rows.Scan(
			&job.ID, &job.Type, &job.Payload, &job.Status,
			&job.CreatedAt, &job.UpdatedAt, &job.StartedAt, &job.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	now := time.Now()
	for i := range jobs {
		_, err := tx.Exec(ctx,
			`UPDATE jobs SET status = $1, started_at = NOW() WHERE id = $2`,
			JobStatusRunning, jobs[i].ID,
		)
		if err != nil {
			return nil, fmt.Errorf("mark job running: %w", err)
		}
		jobs[i].Status = JobStatusRunning
		jobs[i].StartedAt = &now
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return jobs, nil
}

func (s *Store) UpdateJobStatus(ctx context.Context, status JobStatus, id int64) error {
	query :=
		`
			UPDATE jobs
			SET status = $1,
				started_at = CASE WHEN $1 = 'running' THEN NOW() ELSE started_at END,
				completed_at = CASE WHEN $1 IN ('completed', 'failed') THEN NOW() ELSE completed_at END
			WHERE id = $2
		`
	_, err := s.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}

	return nil
}
