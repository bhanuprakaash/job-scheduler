package store

import (
	"context"
	"fmt"
	"math"
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
		SELECT id, type, payload, status, created_at, updated_at, started_at, completed_at, last_err, retry_count
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
			&job.ErrorMessage,
			&job.RetryCount,
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
		WHERE status = $1 AND next_run_at <= NOW()
		ORDER BY next_run_at ASC
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

func (s *Store) HandleJobFailure(ctx context.Context, jobId int64, errMsg string) error {

	query :=
		`
			UPDATE jobs
			SET
				retry_count = retry_count + 1,
				last_err = $2,
				status = CASE 
					WHEN retry_count + 1 >= max_retries THEN 'failed'
					ELSE 'pending'
				END,

				next_run_at = CASE 
					WHEN retry_count + 1 >= max_retries THEN next_run_at
					ELSE NOW() + (POWER(2, retry_count + 1) * INTERVAL '1 second')
				END,

				completed_at = CASE
					WHEN retry_count + 1 >= max_retries THEN NOW()
					ELSE NULL
				END
			WHERE id = $1 AND status = 'running'
		`

	_, err := s.db.Exec(ctx, query, jobId, errMsg)
	if err != nil {
		return fmt.Errorf("handle job failure: %w", err)
	}
	return nil

}

func (s *Store) GetArchivedJobs(ctx context.Context, duration time.Duration, limit int) ([]Job, error) {

	query :=
		`
			SELECT id, type, payload, status, created_at, completed_at
			FROM jobs
			WHERE status = 'completed' AND completed_at < NOW() - $1::INTERVAL
			LIMIT $2
		`

	interval := fmt.Sprintf("%f seconds", duration.Seconds())
	rows, err := s.db.Query(ctx, query, interval, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.CreatedAt, &j.CompletedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}

func (s *Store) BatchDeleteJobs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	query := `DELETE FROM jobs WHERE id = ANY($1)`
	_, err := s.db.Exec(ctx, query, ids)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ListJobs(ctx context.Context, limit, offset int) (*PaginatedJobs, error) {

	var total int64
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM jobs").Scan(&total); err != nil {
		return nil, err
	}

	if total == 0 {
		return &PaginatedJobs{
			Jobs: []Job{},
			Meta: PaginationMetadata{
				CurrentPage:  1,
				TotalPages:   0,
				TotalRecords: 0,
				Limit:        limit,
			},
		}, nil
	}

	query := `
		SELECT id, type, payload, status, created_at, updated_at, started_at, completed_at, last_err, retry_count
		FROM jobs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []Job{}

	for rows.Next() {
		var j Job
		if err := rows.Scan(
			&j.ID,
			&j.Type,
			&j.Payload,
			&j.Status,
			&j.CreatedAt,
			&j.UpdatedAt,
			&j.StartedAt,
			&j.CompletedAt,
			&j.ErrorMessage,
			&j.RetryCount,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	currentPage := (offset / limit) + 1
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &PaginatedJobs{
		Jobs: jobs,
		Meta: PaginationMetadata{
			CurrentPage:  currentPage,
			TotalPages:   totalPages,
			TotalRecords: total,
			Limit:        limit,
		},
	}, nil
}

func (s *Store) GetStats(ctx context.Context) (*JobStats, error) {
	query := `
		SELECT status, COUNT(*)
		FROM jobs
		GROUP BY status
		`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := &JobStats{}
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}

		switch status {
		case "pending":
			stats.Pending = count
		case "running":
			stats.Running = count
		case "completed":
			stats.Completed = count
		case "failed":
			stats.Failed = count
		}
	}

	return stats, nil
}

func (s *Store) RepeatStuckJobs(ctx context.Context, interval time.Duration) (int64, error) {
	query :=
		`	
			UPDATE jobs
			SET status = 'pending',
				updated_at = NOW(),
				last_err = 'job execution timed out (stuck)'
			WHERE 
				status = 'running'
				AND started_at < NOW() - $1::INTERVAL
		`

	intervalStr := fmt.Sprintf("%f seconds", interval.Seconds())
	result, err := s.db.Exec(ctx, query, intervalStr)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil

}
