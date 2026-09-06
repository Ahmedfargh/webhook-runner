package controller

import (
	"context"

	pb "auditService/api/proto/v1"
	"auditService/internal/modules/audit/presenter"
	"auditService/internal/modules/audit/repository"
	"auditService/internal/modules/audit/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuditController struct {
	pb.UnimplementedAuditServiceServer
	service   service.AuditService
	presenter presenter.AuditPresenter
}

func NewAuditController(service service.AuditService, presenter presenter.AuditPresenter) *AuditController {
	return &AuditController{
		service:   service,
		presenter: presenter,
	}
}

func (c *AuditController) RecordAuditLog(ctx context.Context, req *pb.RecordAuditLogRequest) (*pb.RecordAuditLogResponse, error) {
	logModel := c.presenter.FromRecordRequest(req)
	saved, err := c.service.RecordLog(ctx, logModel)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to record audit log: %v", err)
	}

	return &pb.RecordAuditLogResponse{
		Success:  true,
		Message:  "Audit log recorded successfully",
		AuditLog: c.presenter.ToProto(saved),
	}, nil
}

func (c *AuditController) ListAuditLogs(ctx context.Context, req *pb.ListAuditLogsRequest) (*pb.ListAuditLogsResponse, error) {
	filter := repository.AuditFilter{
		ActorID:     req.ActorId,
		ServiceName: req.ServiceName,
		Resource:    req.Resource,
		Action:      req.Action,
		Status:      req.Status,
		Search:      req.Search,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Page:        int(req.Page),
		Limit:       int(req.Limit),
	}

	logs, total, err := c.service.ListLogs(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list audit logs: %v", err)
	}

	return &pb.ListAuditLogsResponse{
		AuditLogs: c.presenter.ToProtoList(logs),
		Total:     total,
		Page:      req.Page,
		Limit:     req.Limit,
	}, nil
}

func (c *AuditController) GetAuditLog(ctx context.Context, req *pb.GetAuditLogRequest) (*pb.GetAuditLogResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid audit log id: %v", err)
	}

	log, err := c.service.GetLog(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "audit log not found: %v", err)
	}

	return &pb.GetAuditLogResponse{
		AuditLog: c.presenter.ToProto(log),
	}, nil
}
