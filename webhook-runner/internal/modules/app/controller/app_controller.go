package controller

import (
	"context"

	pb "webhookRunner/api/proto/v1"
	"webhookRunner/internal/modules/app/presenter"
	"webhookRunner/internal/modules/app/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppController struct {
	pb.UnimplementedAppServiceServer
	service   service.AppService
	presenter presenter.AppPresenter
}

func NewAppController(s service.AppService, p presenter.AppPresenter) *AppController {
	return &AppController{
		service:   s,
		presenter: p,
	}
}

func (c *AppController) CreateApp(ctx context.Context, req *pb.CreateAppRequest) (*pb.CreateAppResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	app, err := c.service.CreateApp(ctx, userID, req.Name, req.WebhookUrl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create app: %v", err)
	}

	return &pb.CreateAppResponse{
		App: c.presenter.ToProto(app),
	}, nil
}

func (c *AppController) GetApp(ctx context.Context, req *pb.GetAppRequest) (*pb.GetAppResponse, error) {
	var id uuid.UUID
	var userID uuid.UUID
	if req.Id != "" {
		id, _ = uuid.Parse(req.Id)
	}
	if req.UserId != "" {
		userID, _ = uuid.Parse(req.UserId)
	}

	app, err := c.service.GetApp(ctx, id, userID, req.AppId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "app not found: %v", err)
	}

	return &pb.GetAppResponse{
		App: c.presenter.ToProto(app),
	}, nil
}

func (c *AppController) ListApps(ctx context.Context, req *pb.ListAppsRequest) (*pb.ListAppsResponse, error) {
	var userID uuid.UUID
	if req.UserId != "" {
		userID, _ = uuid.Parse(req.UserId)
	}

	apps, total, err := c.service.ListApps(ctx, userID, int(req.Page), int(req.Limit), req.Search)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list apps: %v", err)
	}

	return &pb.ListAppsResponse{
		Apps:  c.presenter.ToProtoList(apps),
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

func (c *AppController) UpdateApp(ctx context.Context, req *pb.UpdateAppRequest) (*pb.UpdateAppResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid app id: %v", err)
	}
	var userID uuid.UUID
	if req.UserId != "" {
		userID, _ = uuid.Parse(req.UserId)
	}

	app, err := c.service.UpdateApp(ctx, id, userID, req.Name, req.WebhookUrl, req.IsActive)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update app: %v", err)
	}

	return &pb.UpdateAppResponse{
		App: c.presenter.ToProto(app),
	}, nil
}

func (c *AppController) DeleteApp(ctx context.Context, req *pb.DeleteAppRequest) (*pb.DeleteAppResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid app id: %v", err)
	}
	var userID uuid.UUID
	if req.UserId != "" {
		userID, _ = uuid.Parse(req.UserId)
	}

	if err := c.service.DeleteApp(ctx, id, userID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete app: %v", err)
	}

	return &pb.DeleteAppResponse{
		Success: true,
		Message: "App deleted successfully",
	}, nil
}

func (c *AppController) RotateSecrets(ctx context.Context, req *pb.RotateSecretsRequest) (*pb.RotateSecretsResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid app id: %v", err)
	}
	var userID uuid.UUID
	if req.UserId != "" {
		userID, _ = uuid.Parse(req.UserId)
	}

	app, err := c.service.RotateSecrets(ctx, id, userID, req.RotateAppSecret, req.RotateWebhookSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to rotate secrets: %v", err)
	}

	return &pb.RotateSecretsResponse{
		App: c.presenter.ToProto(app),
	}, nil
}
