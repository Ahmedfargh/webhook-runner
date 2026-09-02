package controller

import (
	"context"

	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/modules/admin/presenter"
	"accounts/internal/modules/admin/service"
	"accounts/internal/pkg/grpcerr"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminController implements accountsv1.AdminServiceServer
type AdminController struct {
	accountsv1.UnimplementedAdminServiceServer
	service   service.AdminService
	presenter presenter.AdminPresenter
}

// NewAdminController creates a new AdminController instance
func NewAdminController(svc service.AdminService, pres presenter.AdminPresenter) *AdminController {
	return &AdminController{
		service:   svc,
		presenter: pres,
	}
}

func (c *AdminController) CreateAdmin(ctx context.Context, req *accountsv1.CreateAdminRequest) (*accountsv1.AdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	input := service.CreateAdminInput{
		Name:          req.GetName(),
		Email:         req.GetEmail(),
		Phone:         req.GetPhone(),
		Password:      req.GetPassword(),
		CountryID:     req.GetCountryId(),
		RoleIDs:       req.GetRoleIds(),
		PermissionIDs: req.GetPermissionIds(),
	}

	admin, err := c.service.CreateAdmin(ctx, input)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(admin), nil
}

func (c *AdminController) GetAdmin(ctx context.Context, req *accountsv1.GetAdminRequest) (*accountsv1.AdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid admin UUID format")
	}

	admin, err := c.service.GetAdmin(ctx, id)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(admin), nil
}

func (c *AdminController) UpdateAdmin(ctx context.Context, req *accountsv1.UpdateAdminRequest) (*accountsv1.AdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid admin UUID format")
	}

	var passwordPtr *string
	if req.Password != nil {
		pwd := req.GetPassword()
		passwordPtr = &pwd
	}

	input := service.UpdateAdminInput{
		ID:            id,
		Name:          req.GetName(),
		Email:         req.GetEmail(),
		Phone:         req.GetPhone(),
		Password:      passwordPtr,
		CountryID:     req.GetCountryId(),
		RoleIDs:       req.GetRoleIds(),
		PermissionIDs: req.GetPermissionIds(),
	}

	admin, err := c.service.UpdateAdmin(ctx, input)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(admin), nil
}

func (c *AdminController) DeleteAdmin(ctx context.Context, req *accountsv1.DeleteAdminRequest) (*accountsv1.DeleteAdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid admin UUID format")
	}

	if err := c.service.DeleteAdmin(ctx, id); err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return &accountsv1.DeleteAdminResponse{
		Success: true,
		Message: "Admin deleted successfully",
	}, nil
}

func (c *AdminController) ListAdmins(ctx context.Context, req *accountsv1.ListAdminsRequest) (*accountsv1.ListAdminsResponse, error) {
	var page, pageSize int
	var search string

	if req != nil && req.GetPagination() != nil {
		page = int(req.GetPagination().GetPage())
		pageSize = int(req.GetPagination().GetPageSize())
		search = req.GetPagination().GetSearch()
	}

	admins, total, err := c.service.ListAdmins(ctx, page, pageSize, search)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToListProto(admins, total, page, pageSize), nil
}

func (c *AdminController) AssignRolesToAdmin(ctx context.Context, req *accountsv1.AssignRolesToAdminRequest) (*accountsv1.AdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	adminID, err := uuid.Parse(req.GetAdminId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid admin UUID format")
	}

	admin, err := c.service.AssignRoles(ctx, adminID, req.GetRoleIds())
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(admin), nil
}

func (c *AdminController) AssignPermissionsToAdmin(ctx context.Context, req *accountsv1.AssignPermissionsToAdminRequest) (*accountsv1.AdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	adminID, err := uuid.Parse(req.GetAdminId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid admin UUID format")
	}

	admin, err := c.service.AssignPermissions(ctx, adminID, req.GetPermissionIds())
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(admin), nil
}
