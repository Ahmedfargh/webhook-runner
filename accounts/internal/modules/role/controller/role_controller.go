package controller

import (
	"context"

	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/modules/role/presenter"
	"accounts/internal/modules/role/service"
	"accounts/internal/pkg/grpcerr"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RoleController implements accountsv1.RoleServiceServer
type RoleController struct {
	accountsv1.UnimplementedRoleServiceServer
	service   service.RoleService
	presenter presenter.RolePresenter
}

// NewRoleController creates a new RoleController
func NewRoleController(svc service.RoleService, pres presenter.RolePresenter) *RoleController {
	return &RoleController{
		service:   svc,
		presenter: pres,
	}
}

func (c *RoleController) CreateRole(ctx context.Context, req *accountsv1.CreateRoleRequest) (*accountsv1.RoleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	role, err := c.service.CreateRole(ctx, req.GetName(), req.GetPermissionIds())
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(role), nil
}

func (c *RoleController) GetRole(ctx context.Context, req *accountsv1.GetRoleRequest) (*accountsv1.RoleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role UUID format")
	}

	role, err := c.service.GetRole(ctx, id)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(role), nil
}

func (c *RoleController) UpdateRole(ctx context.Context, req *accountsv1.UpdateRoleRequest) (*accountsv1.RoleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role UUID format")
	}

	role, err := c.service.UpdateRole(ctx, id, req.GetName(), req.GetPermissionIds())
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(role), nil
}

func (c *RoleController) DeleteRole(ctx context.Context, req *accountsv1.DeleteRoleRequest) (*accountsv1.DeleteRoleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role UUID format")
	}

	if err := c.service.DeleteRole(ctx, id); err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return &accountsv1.DeleteRoleResponse{
		Success: true,
		Message: "Role deleted successfully",
	}, nil
}

func (c *RoleController) ListRoles(ctx context.Context, req *accountsv1.ListRolesRequest) (*accountsv1.ListRolesResponse, error) {
	var page, pageSize int
	var search string

	if req != nil && req.GetPagination() != nil {
		page = int(req.GetPagination().GetPage())
		pageSize = int(req.GetPagination().GetPageSize())
		search = req.GetPagination().GetSearch()
	}

	roles, total, err := c.service.ListRoles(ctx, page, pageSize, search)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToListProto(roles, total, page, pageSize), nil
}

func (c *RoleController) AssignPermissionsToRole(ctx context.Context, req *accountsv1.AssignPermissionsToRoleRequest) (*accountsv1.RoleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role UUID format")
	}

	role, err := c.service.AssignPermissions(ctx, id, req.GetPermissionIds())
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(role), nil
}
