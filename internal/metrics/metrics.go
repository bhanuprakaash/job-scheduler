package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)


var (
	JobsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "job_scheduler_jobs_processed_total",
			Help: "The total number of jobs processed by the workers",
		},
		[]string{"job_type", "status"},
	)


	JobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "job_scheduler_job_duration_seconds",
			Help:    "Duration of job execution in seconds",
			Buckets: prometheus.DefBuckets, // Default buckets (0.1s, 0.5s, 1s, etc.)
		},
		[]string{"job_type"},
	)

	ActiveWorkers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "job_scheduler_active_workers",
			Help: "Current number of workers processing jobs",
		},
	)
)