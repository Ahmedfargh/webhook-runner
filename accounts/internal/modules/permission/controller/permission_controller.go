package controller

import (
	"context"

	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/modules/permission/presenter"
	"accounts/internal/modules/permission/service"
	"accounts/internal/pkg/grpcerr"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PermissionController implements accountsv1.PermissionServiceServer
type PermissionController struct {
	accountsv1.UnimplementedPermissionServiceServer
	service   service.PermissionService
	presenter presenter.PermissionPresenter
}

// NewPermissionController creates a new PermissionController instance
func NewPermissionController(svc service.PermissionService, pres presenter.PermissionPresenter) *PermissionController {
	return &PermissionController{
		service:   svc,
		presenter: pres,
	}
}

func (c *PermissionController) CreatePermission(ctx context.Context, req *accountsv1.CreatePermissionRequest) (*accountsv1.PermissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	permission, err := c.service.CreatePermission(ctx, req.GetName())
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(permission), nil
}

func (c *PermissionController) GetPermission(ctx context.Context, req *accountsv1.GetPermissionRequest) (*accountsv1.PermissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid permission UUID format")
	}

	permission, err := c.service.GetPermission(ctx, id)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(permission), nil
}

func (c *PermissionController) UpdatePermission(ctx context.Context, req *accountsv1.UpdatePermissionRequest) (*accountsv1.PermissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid permission UUID format")
	}

	permission, err := c.service.UpdatePermission(ctx, id, req.GetName())
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(permission), nil
}

func (c *PermissionController) DeletePermission(ctx context.Context, req *accountsv1.DeletePermissionRequest) (*accountsv1.DeletePermissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid permission UUID format")
	}

	if err := c.service.DeletePermission(ctx, id); err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return &accountsv1.DeletePermissionResponse{
		Success: true,
		Message: "Permission deleted successfully",
	}, nil
}

func (c *PermissionController) ListPermissions(ctx context.Context, req *accountsv1.ListPermissionsRequest) (*accountsv1.ListPermissionsResponse, error) {
	var page, pageSize int
	var search string

	if req != nil && req.GetPagination() != nil {
		page = int(req.GetPagination().GetPage())
		pageSize = int(req.GetPagination().GetPageSize())
		search = req.GetPagination().GetSearch()
	}

	permissions, total, err := c.service.ListPermissions(ctx, page, pageSize, search)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToListProto(permissions, total, page, pageSize), nil
}
