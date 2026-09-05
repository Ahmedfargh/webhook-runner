package controller

import (
	"context"
	"errors"

	subscriptionsv1 "subscriptions/api/proto/v1"
	"subscriptions/internal/modules/manual_payment/presenter"
	"subscriptions/internal/modules/manual_payment/service"
	"subscriptions/internal/pkg/apperrors"
	"subscriptions/internal/pkg/uuidutil"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ManualPaymentController struct {
	subscriptionsv1.UnimplementedManualPaymentServiceServer
	service   service.ManualPaymentService
	presenter presenter.ManualPaymentPresenter
}

func NewManualPaymentController(service service.ManualPaymentService, presenter presenter.ManualPaymentPresenter) *ManualPaymentController {
	return &ManualPaymentController{
		service:   service,
		presenter: presenter,
	}
}

func (c *ManualPaymentController) SubmitManualPayment(ctx context.Context, req *subscriptionsv1.SubmitManualPaymentRequest) (*subscriptionsv1.ManualPaymentResponse, error) {
	invoiceID, err := uuid.Parse(req.GetInvoiceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid invoice ID")
	}

	userID := uuidutil.ParseOrHash(req.GetUserId())

	payment, err := c.service.SubmitPayment(ctx, service.SubmitPaymentInput{
		InvoiceID:            invoiceID,
		UserID:               userID,
		Amount:               req.GetAmount(),
		Currency:             req.GetCurrency(),
		PaymentMethod:        req.GetPaymentMethod(),
		TransactionReference: req.GetTransactionReference(),
		PayerName:            req.GetPayerName(),
		PayerNotes:           req.GetPayerNotes(),
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, apperrors.ErrInvoiceAlreadyPaid) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return c.presenter.ToProto(payment), nil
}

func (c *ManualPaymentController) ReviewManualPayment(ctx context.Context, req *subscriptionsv1.ReviewManualPaymentRequest) (*subscriptionsv1.ReviewManualPaymentResponse, error) {
	paymentID, err := uuid.Parse(req.GetPaymentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment ID")
	}

	var adminID *uuid.UUID
	if req.GetAdminId() != "" {
		if aid, err := uuid.Parse(req.GetAdminId()); err == nil {
			adminID = &aid
		}
	}

	payment, err := c.service.ReviewPayment(ctx, service.ReviewPaymentInput{
		PaymentID:  paymentID,
		Approve:    req.GetApprove(),
		AdminNotes: req.GetAdminNotes(),
		AdminID:    adminID,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	action := "rejected"
	if req.GetApprove() {
		action = "approved and subscription activated"
	}

	return &subscriptionsv1.ReviewManualPaymentResponse{
		Success: true,
		Message: "Payment record " + action + " successfully",
		Payment: c.presenter.ToProto(payment),
	}, nil
}

func (c *ManualPaymentController) ListManualPayments(ctx context.Context, req *subscriptionsv1.ListManualPaymentsRequest) (*subscriptionsv1.ListManualPaymentsResponse, error) {
	payments, total, err := c.service.ListPayments(ctx, int(req.GetPage()), int(req.GetPageSize()), req.GetStatus(), req.GetSearch())
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

	return c.presenter.ToListProto(payments, total, page, pageSize), nil
}
