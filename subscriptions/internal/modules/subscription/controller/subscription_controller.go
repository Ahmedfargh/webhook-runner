package controller

import (
	"context"
	"errors"
	"time"

	subscriptionsv1 "subscriptions/api/proto/v1"
	"subscriptions/internal/models"
	"subscriptions/internal/modules/subscription/presenter"
	"subscriptions/internal/modules/subscription/service"
	"subscriptions/internal/pkg/apperrors"
	"subscriptions/internal/pkg/uuidutil"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubscriptionController struct {
	subscriptionsv1.UnimplementedSubscriptionServiceServer
	service   service.SubscriptionService
	presenter presenter.SubscriptionPresenter
}

func NewSubscriptionController(service service.SubscriptionService, presenter presenter.SubscriptionPresenter) *SubscriptionController {
	return &SubscriptionController{
		service:   service,
		presenter: presenter,
	}
}

func (c *SubscriptionController) Subscribe(ctx context.Context, req *subscriptionsv1.SubscribeRequest) (*subscriptionsv1.SubscribeResponse, error) {
	userID := uuidutil.ParseOrHash(req.GetUserId())
	if userID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	planID, err := uuid.Parse(req.GetPlanId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid plan ID")
	}

	res, err := c.service.Subscribe(ctx, service.SubscribeInput{
		UserID:        userID,
		PlanID:        planID,
		BillingCycle:  models.BillingCycle(req.GetBillingCycle()),
		PaymentMethod: req.GetPaymentMethod(),
		Notes:         req.GetNotes(),
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &subscriptionsv1.SubscribeResponse{
		Subscription:        c.presenter.ToProto(res.Subscription),
		InvoiceId:           res.InvoiceID,
		InvoiceNumber:       res.InvoiceNumber,
		AmountDue:           res.AmountDue,
		Currency:            res.Currency,
		PaymentInstructions: res.PaymentInstructions,
	}, nil
}

func (c *SubscriptionController) GetUserSubscription(ctx context.Context, req *subscriptionsv1.GetUserSubscriptionRequest) (*subscriptionsv1.SubscriptionResponse, error) {
	userID := uuidutil.ParseOrHash(req.GetUserId())
	if userID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	sub, err := c.service.GetUserSubscription(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return c.presenter.ToProto(sub), nil
}

func (c *SubscriptionController) CancelSubscription(ctx context.Context, req *subscriptionsv1.CancelSubscriptionRequest) (*subscriptionsv1.CancelSubscriptionResponse, error) {
	userID := uuidutil.ParseOrHash(req.GetUserId())
	if userID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	sub, err := c.service.CancelSubscription(ctx, userID, req.GetReason(), req.GetImmediately())
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &subscriptionsv1.CancelSubscriptionResponse{
		Success:      true,
		Message:      "Subscription cancellation processed",
		Subscription: c.presenter.ToProto(sub),
	}, nil
}

func (c *SubscriptionController) AdminOverrideSubscription(ctx context.Context, req *subscriptionsv1.AdminOverrideSubscriptionRequest) (*subscriptionsv1.SubscriptionResponse, error) {
	userID := uuidutil.ParseOrHash(req.GetUserId())
	if userID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	var planID uuid.UUID
	if req.GetPlanId() != "" {
		planID, _ = uuid.Parse(req.GetPlanId())
	}

	var periodEnd *time.Time
	if req.GetCurrentPeriodEnd() != "" {
		if t, err := time.Parse(time.RFC3339, req.GetCurrentPeriodEnd()); err == nil {
			periodEnd = &t
		}
	}

	sub, err := c.service.AdminOverrideSubscription(ctx, service.AdminOverrideInput{
		UserID:           userID,
		PlanID:           planID,
		Status:           models.SubscriptionStatus(req.GetStatus()),
		CurrentPeriodEnd: periodEnd,
		AdminNotes:       req.GetAdminNotes(),
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return c.presenter.ToProto(sub), nil
}

func (c *SubscriptionController) ListSubscriptions(ctx context.Context, req *subscriptionsv1.ListSubscriptionsRequest) (*subscriptionsv1.ListSubscriptionsResponse, error) {
	subs, total, err := c.service.ListSubscriptions(ctx, int(req.GetPage()), int(req.GetPageSize()), req.GetStatus(), req.GetSearch())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 10
	}

	return c.presenter.ToListProto(subs, total, page, pageSize), nil
}
