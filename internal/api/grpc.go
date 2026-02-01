package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
	"github.com/bhanuprakaash/job-scheduler/internal/worker"
	pb "github.com/bhanuprakaash/job-scheduler/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedJobSchedulerServer
	store    store.Storer
	registry *worker.Registry
}

func NewServer(store store.Storer, registry *worker.Registry) *Server {
	return &Server{
		store:    store,
		registry: registry,
	}
}

func (s *Server) SubmitJob(ctx context.Context, req *pb.SubmitJobRequest) (*pb.SubmitJobResponse, error) {
	logger.Info("Received job submission", "type", req.Type, "payload", req.Payload)

	if req.Type == "" {
		return nil, fmt.Errorf("job type is required")
	}
	if !s.registry.Has(req.Type) {
		logger.Error("Invalid job type submitted", "type", req.Type)
		return nil, status.Errorf(codes.InvalidArgument, "job type '%s' is not registered", req.Type)
	}

	if req.Payload == "" {
		req.Payload = "{}"
	}

	job, err := s.store.CreateJob(ctx, req.Type, req.Payload)
	if err != nil {
		logger.Error("Failed to create job", "error", err)
		return nil, err
	}

	logger.Info("job created successfully", "job_id", job.ID)

	return &pb.SubmitJobResponse{
		JobId:  strconv.FormatInt(job.ID, 10),
		Status: string(job.Status),
	}, nil
}

func (s *Server) GetJob(ctx context.Context, req *pb.GetJobRequest) (*pb.GetJobResponse, error) {

	id, err := strconv.ParseInt(req.JobId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid job id format: %v", req.JobId)
	}

	job, err := s.store.GetJobByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get job", "error", err)
		return nil, err
	}

	resp := &pb.GetJobResponse{
		JobId:      strconv.FormatInt(job.ID, 10),
		Type:       job.Type,
		Payload:    job.Payload,
		Status:     string(job.Status),
		CreatedAt:  job.CreatedAt.Format("2006-01-02T15:04:05Z"),
		RetryCount: strconv.Itoa(job.RetryCount),
	}

	if job.ErrorMessage.Valid {
		resp.ErrorMessage = job.ErrorMessage.String
	}

	if job.CompletedAt != nil {
		resp.CompletedAt = job.CompletedAt.Format("2006-01-02T15:04:05Z")
	}

	return resp, nil

}

func (s *Server) ListJobs(ctx context.Context, req *pb.ListJobRequest) (*pb.ListJobResponse, error) {

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	jobs, err := s.store.ListJobs(ctx, int(limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list jobs: %v", err)
	}

	var pbJobs []*pb.GetJobResponse
	for _, j := range jobs.Jobs {
		jobResp := &pb.GetJobResponse{
			JobId:        fmt.Sprintf("%d", j.ID),
			Type:         j.Type,
			Payload:      j.Payload,
			Status:       string(j.Status),
			CreatedAt:    j.CreatedAt.Format("2006-01-02T15:04:05Z"),
			ErrorMessage: j.ErrorMessage.String,
			RetryCount:   strconv.Itoa(j.RetryCount),
		}

		if j.CompletedAt != nil {
			jobResp.CompletedAt = j.CompletedAt.Format("2006-01-02T15:04:05Z")
		}

		pbJobs = append(pbJobs, jobResp)
	}

	var pagination *pb.PaginationMetaData
	pagination = &pb.PaginationMetaData{
		CurrentPage:  int32(jobs.Meta.CurrentPage),
		TotalPages:   int32(jobs.Meta.TotalPages),
		TotalRecords: jobs.Meta.TotalRecords,
		Limit:        int32(jobs.Meta.Limit),
	}

	return &pb.ListJobResponse{
		Jobs: pbJobs,
		Meta: pagination,
	}, nil

}

func (s *Server) GetJobStats(ctx context.Context, req *pb.GetJobStatsRequest) (*pb.GetJobStatusResponse, error) {
	stats, err := s.store.GetStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch stats %v", err)
	}

	return &pb.GetJobStatusResponse{
		PendingJobs:   stats.Pending,
		RunningJobs:   stats.Running,
		CompletedJobs: stats.Completed,
		FailedJobs:    stats.Failed,
		TotalJobs:     stats.Pending + stats.Running + stats.Completed + stats.Failed,
	}, nil
}
