package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
	pb "github.com/bhanuprakaash/job-scheduler/proto"
)

type Server struct {
	pb.UnimplementedJobSchedulerServer
	store *store.Store
}

func NewServer(store *store.Store) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) SubmitJob(ctx context.Context, req *pb.SubmitJobRequest) (*pb.SubmitJobResponse, error) {
	logger.Info("Received job submission", "type", req.Type, "payload", req.Payload)

	if req.Type == "" {
		return nil, fmt.Errorf("job type is required")
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
		JobId:     strconv.FormatInt(job.ID, 10),
		Type:      job.Type,
		Payload:   job.Payload,
		Status:    string(job.Status),
		CreatedAt: job.CreatedAt.Format("2006-01-02T15:04:05Z"),
		ErrorMessage: job.ErrorMessage,
	}

	if job.CompletedAt != nil {
		resp.CompletedAt = job.CompletedAt.Format("2006-01-02T15:04:05Z")
	}

	return resp, nil

}
