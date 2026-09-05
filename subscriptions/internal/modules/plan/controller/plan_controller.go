package controller

import (
	"context"
	"errors"

	subscriptionsv1 "subscriptions/api/proto/v1"
	"subscriptions/internal/modules/plan/presenter"
	"subscriptions/internal/modules/plan/service"
	"subscriptions/internal/pkg/apperrors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PlanController struct {
	subscriptionsv1.UnimplementedPlanServiceServer
	service   service.PlanService
	presenter presenter.PlanPresenter
}

func NewPlanController(service service.PlanService, presenter presenter.PlanPresenter) *PlanController {
	return &PlanController{
		service:   service,
		presenter: presenter,
	}
}

func (c *PlanController) CreatePlan(ctx context.Context, req *subscriptionsv1.CreatePlanRequest) (*subscriptionsv1.PlanResponse, error) {
	plan, err := c.service.CreatePlan(ctx, service.CreatePlanInput{
		Name:              req.GetName(),
		Code:              req.GetCode(),
		Description:       req.GetDescription(),
		PriceMonthly:      req.GetPriceMonthly(),
		PriceAnnually:     req.GetPriceAnnually(),
		Currency:          req.GetCurrency(),
		MaxWebhooks:       req.GetMaxWebhooks(),
		MaxEventsPerMonth: req.GetMaxEventsPerMonth(),
		MaxTeamMembers:    req.GetMaxTeamMembers(),
		Features:          req.GetFeatures(),
		IsActive:          req.GetIsActive(),
		IsPopular:         req.GetIsPopular(),
		TierLevel:         req.GetTierLevel(),
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, apperrors.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return c.presenter.ToProto(plan), nil
}

func (c *PlanController) GetPlan(ctx context.Context, req *subscriptionsv1.GetPlanRequest) (*subscriptionsv1.PlanResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid plan ID format")
	}
	plan, err := c.service.GetPlan(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return c.presenter.ToProto(plan), nil
}

func (c *PlanController) UpdatePlan(ctx context.Context, req *subscriptionsv1.UpdatePlanRequest) (*subscriptionsv1.PlanResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid plan ID format")
	}
	plan, err := c.service.UpdatePlan(ctx, service.UpdatePlanInput{
		ID:                id,
		Name:              req.GetName(),
		Description:       req.GetDescription(),
		PriceMonthly:      req.GetPriceMonthly(),
		PriceAnnually:     req.GetPriceAnnually(),
		Currency:          req.GetCurrency(),
		MaxWebhooks:       req.GetMaxWebhooks(),
		MaxEventsPerMonth: req.GetMaxEventsPerMonth(),
		MaxTeamMembers:    req.GetMaxTeamMembers(),
		Features:          req.GetFeatures(),
		IsActive:          req.GetIsActive(),
		IsPopular:         req.GetIsPopular(),
		TierLevel:         req.GetTierLevel(),
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return c.presenter.ToProto(plan), nil
}

func (c *PlanController) DeletePlan(ctx context.Context, req *subscriptionsv1.DeletePlanRequest) (*subscriptionsv1.DeletePlanResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid plan ID format")
	}
	if err := c.service.DeletePlan(ctx, id); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &subscriptionsv1.DeletePlanResponse{
		Success: true,
		Message: "Plan deleted successfully",
	}, nil
}

func (c *PlanController) ListPlans(ctx context.Context, req *subscriptionsv1.ListPlansRequest) (*subscriptionsv1.ListPlansResponse, error) {
	plans, err := c.service.ListPlans(ctx, req.GetIncludeInactive())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return c.presenter.ToListProto(plans), nil
}
