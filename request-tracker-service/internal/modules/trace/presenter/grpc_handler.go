package presenter

import (
	"context"
	"time"

	pb "requestTrackerService/api/proto/v1"
	"requestTrackerService/internal/models"
	"requestTrackerService/internal/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedRequestTrackerServiceServer
	repo repository.TraceRepository
}

func NewGRPCHandler(repo repository.TraceRepository) *GRPCHandler {
	return &GRPCHandler{repo: repo}
}

func (h *GRPCHandler) RecordTrace(ctx context.Context, req *pb.RecordTraceRequest) (*pb.RecordTraceResponse, error) {
	if req == nil || req.Trace == nil {
		return nil, status.Error(codes.InvalidArgument, "trace payload is required")
	}

	t := req.Trace
	var traceID uuid.UUID
	if t.Id != "" {
		if parsed, err := uuid.Parse(t.Id); err == nil {
			traceID = parsed
		} else {
			traceID = uuid.New()
		}
	} else {
		traceID = uuid.New()
	}

	receivedAt := time.Now().UTC()
	if t.ReceivedAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, t.ReceivedAt); err == nil {
			receivedAt = parsed
		} else if parsed, err := time.Parse(time.RFC3339, t.ReceivedAt); err == nil {
			receivedAt = parsed
		}
	}

	completedAt := time.Now().UTC()
	if t.CompletedAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, t.CompletedAt); err == nil {
			completedAt = parsed
		} else if parsed, err := time.Parse(time.RFC3339, t.CompletedAt); err == nil {
			completedAt = parsed
		}
	}

	model := &models.RequestTrace{
		ID:           traceID,
		TraceID:      t.TraceId,
		RequestID:    t.RequestId,
		ActorType:    t.ActorType,
		ActorID:      t.ActorId,
		ActorName:    t.ActorName,
		ActorEmail:   t.ActorEmail,
		ActorRole:    t.ActorRole,
		ServiceName:  t.ServiceName,
		Method:       t.Method,
		Path:         t.Path,
		Route:        t.Route,
		QueryParams:  t.QueryParams,
		ClientIP:     t.ClientIp,
		UserAgent:    t.UserAgent,
		StatusCode:   int(t.StatusCode),
		LifetimeMs:   t.LifetimeMs,
		RequestBody:  t.RequestBody,
		ResponseBody: t.ResponseBody,
		ErrorMessage: t.ErrorMessage,
		SpansJSON:    t.SpansJson,
		ReceivedAt:   receivedAt,
		CompletedAt:  completedAt,
		CreatedAt:    time.Now().UTC(),
	}

	if err := h.repo.Create(ctx, model); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store trace: %v", err)
	}

	return &pb.RecordTraceResponse{
		Success: true,
		Message: "Trace recorded successfully",
		Trace:   toPB(model),
	}, nil
}

func (h *GRPCHandler) ListTraces(ctx context.Context, req *pb.ListTracesRequest) (*pb.ListTracesResponse, error) {
	filter := repository.TraceFilter{
		ActorType:     req.ActorType,
		ActorID:       req.ActorId,
		Method:        req.Method,
		Route:         req.Route,
		StatusCode:    int(req.StatusCode),
		MinLifetimeMs: req.MinLifetimeMs,
		Search:        req.Search,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Page:          int(req.Page),
		Limit:         int(req.Limit),
	}

	traces, total, err := h.repo.List(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query traces: %v", err)
	}

	pbTraces := make([]*pb.RequestTrace, 0, len(traces))
	for _, t := range traces {
		pbTraces = append(pbTraces, toPB(t))
	}

	return &pb.ListTracesResponse{
		Traces: pbTraces,
		Total:  total,
		Page:   req.Page,
		Limit:  req.Limit,
	}, nil
}

func (h *GRPCHandler) GetTrace(ctx context.Context, req *pb.GetTraceRequest) (*pb.GetTraceResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id or trace_id is required")
	}

	trace, err := h.repo.FindByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "trace not found: %v", err)
	}

	return &pb.GetTraceResponse{
		Trace: toPB(trace),
	}, nil
}

func (h *GRPCHandler) GetTraceStats(ctx context.Context, req *pb.GetTraceStatsRequest) (*pb.GetTraceStatsResponse, error) {
	stats, err := h.repo.GetStats(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate trace stats: %v", err)
	}

	return &pb.GetTraceStatsResponse{
		TotalRequests: stats.TotalRequests,
		AvgLifetimeMs: stats.AvgLifetimeMs,
		P95LifetimeMs: stats.P95LifetimeMs,
		P99LifetimeMs: stats.P99LifetimeMs,
		ErrorCount:    stats.ErrorCount,
		ErrorRate:     stats.ErrorRate,
	}, nil
}

func toPB(m *models.RequestTrace) *pb.RequestTrace {
	if m == nil {
		return nil
	}
	return &pb.RequestTrace{
		Id:           m.ID.String(),
		TraceId:      m.TraceID,
		RequestId:    m.RequestID,
		ActorType:    m.ActorType,
		ActorId:      m.ActorID,
		ActorName:    m.ActorName,
		ActorEmail:   m.ActorEmail,
		ActorRole:    m.ActorRole,
		ServiceName:  m.ServiceName,
		Method:       m.Method,
		Path:         m.Path,
		Route:        m.Route,
		QueryParams:  m.QueryParams,
		ClientIp:     m.ClientIP,
		UserAgent:    m.UserAgent,
		StatusCode:   int32(m.StatusCode),
		LifetimeMs:   m.LifetimeMs,
		RequestBody:  m.RequestBody,
		ResponseBody: m.ResponseBody,
		ErrorMessage: m.ErrorMessage,
		SpansJson:    m.SpansJSON,
		ReceivedAt:   m.ReceivedAt.Format(time.RFC3339Nano),
		CompletedAt:  m.CompletedAt.Format(time.RFC3339Nano),
	}
}
