package controller

import (
	"context"

	pb "webhookRunner/api/proto/v1"
	"webhookRunner/internal/modules/webhook/presenter"
	"webhookRunner/internal/modules/webhook/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WebhookController struct {
	pb.UnimplementedWebhookRunnerServiceServer
	service   service.WebhookService
	presenter presenter.WebhookPresenter
}

func NewWebhookController(s service.WebhookService, p presenter.WebhookPresenter) *WebhookController {
	return &WebhookController{
		service:   s,
		presenter: p,
	}
}

func (c *WebhookController) SendWebhook(ctx context.Context, req *pb.SendWebhookRequest) (*pb.SendWebhookResponse, error) {
	call, err := c.service.SendWebhook(ctx, req.AppId, req.EventName, req.PayloadJson, req.CustomHeaders, req.TargetUrlOverride)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send webhook: %v", err)
	}

	return &pb.SendWebhookResponse{
		WebhookCall: c.presenter.ToProto(call),
		Success:     call.Status == "success",
		Message:     "Webhook dispatched successfully",
	}, nil
}

func (c *WebhookController) ListWebhookCalls(ctx context.Context, req *pb.ListWebhookCallsRequest) (*pb.ListWebhookCallsResponse, error) {
	var userID uuid.UUID
	if req.UserId != "" {
		userID, _ = uuid.Parse(req.UserId)
	}

	calls, total, err := c.service.ListWebhookCalls(ctx, userID, req.AppId, req.Status, int(req.Page), int(req.Limit), req.Search)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list webhook calls: %v", err)
	}

	return &pb.ListWebhookCallsResponse{
		WebhookCalls: c.presenter.ToProtoList(calls),
		Total:        total,
		Page:         req.Page,
		Limit:        req.Limit,
	}, nil
}

func (c *WebhookController) GetWebhookCall(ctx context.Context, req *pb.GetWebhookCallRequest) (*pb.GetWebhookCallResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid webhook call id: %v", err)
	}
	var userID uuid.UUID
	if req.UserId != "" {
		userID, _ = uuid.Parse(req.UserId)
	}

	call, err := c.service.GetWebhookCall(ctx, id, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "webhook call not found: %v", err)
	}

	return &pb.GetWebhookCallResponse{
		WebhookCall: c.presenter.ToProto(call),
	}, nil
}

func (c *WebhookController) RetryWebhookCall(ctx context.Context, req *pb.RetryWebhookCallRequest) (*pb.RetryWebhookCallResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid webhook call id: %v", err)
	}
	var userID uuid.UUID
	if req.UserId != "" {
		userID, _ = uuid.Parse(req.UserId)
	}

	call, err := c.service.RetryWebhookCall(ctx, id, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retry webhook call: %v", err)
	}

	return &pb.RetryWebhookCallResponse{
		WebhookCall: c.presenter.ToProto(call),
		Success:     call.Status == "success",
		Message:     "Webhook retried successfully",
	}, nil
}
