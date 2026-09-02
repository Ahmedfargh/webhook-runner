package controller

import (
	"context"

	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/modules/user/presenter"
	"accounts/internal/modules/user/service"
	"accounts/internal/pkg/grpcerr"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserController implements accountsv1.UserServiceServer
type UserController struct {
	accountsv1.UnimplementedUserServiceServer
	service   service.UserService
	presenter presenter.UserPresenter
}

// NewUserController creates a new UserController instance
func NewUserController(svc service.UserService, pres presenter.UserPresenter) *UserController {
	return &UserController{
		service:   svc,
		presenter: pres,
	}
}

func (c *UserController) CreateUser(ctx context.Context, req *accountsv1.CreateUserRequest) (*accountsv1.UserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	input := service.CreateUserInput{
		Name:      req.GetName(),
		Email:     req.GetEmail(),
		Phone:     req.GetPhone(),
		Password:  req.GetPassword(),
		CountryID: req.GetCountryId(),
	}

	user, err := c.service.CreateUser(ctx, input)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(user), nil
}

func (c *UserController) GetUser(ctx context.Context, req *accountsv1.GetUserRequest) (*accountsv1.UserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user UUID format")
	}

	user, err := c.service.GetUser(ctx, id)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(user), nil
}

func (c *UserController) UpdateUser(ctx context.Context, req *accountsv1.UpdateUserRequest) (*accountsv1.UserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user UUID format")
	}

	var passwordPtr *string
	if req.Password != nil {
		pwd := req.GetPassword()
		passwordPtr = &pwd
	}

	input := service.UpdateUserInput{
		ID:        id,
		Name:      req.GetName(),
		Email:     req.GetEmail(),
		Phone:     req.GetPhone(),
		Password:  passwordPtr,
		CountryID: req.GetCountryId(),
	}

	user, err := c.service.UpdateUser(ctx, input)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToProto(user), nil
}

func (c *UserController) DeleteUser(ctx context.Context, req *accountsv1.DeleteUserRequest) (*accountsv1.DeleteUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user UUID format")
	}

	if err := c.service.DeleteUser(ctx, id); err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return &accountsv1.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

func (c *UserController) ListUsers(ctx context.Context, req *accountsv1.ListUsersRequest) (*accountsv1.ListUsersResponse, error) {
	var page, pageSize int
	var search string

	if req != nil && req.GetPagination() != nil {
		page = int(req.GetPagination().GetPage())
		pageSize = int(req.GetPagination().GetPageSize())
		search = req.GetPagination().GetSearch()
	}

	users, total, err := c.service.ListUsers(ctx, page, pageSize, search)
	if err != nil {
		return nil, grpcerr.ToGRPCError(err)
	}

	return c.presenter.ToListProto(users, total, page, pageSize), nil
}
