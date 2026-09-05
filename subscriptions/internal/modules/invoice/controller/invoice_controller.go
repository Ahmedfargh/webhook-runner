package controller

import (
	"context"
	"errors"
	"time"

	subscriptionsv1 "subscriptions/api/proto/v1"
	"subscriptions/internal/modules/invoice/presenter"
	"subscriptions/internal/modules/invoice/service"
	"subscriptions/internal/pkg/apperrors"
	"subscriptions/internal/pkg/uuidutil"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InvoiceController struct {
	subscriptionsv1.UnimplementedInvoiceServiceServer
	service   service.InvoiceService
	presenter presenter.InvoicePresenter
}

func NewInvoiceController(service service.InvoiceService, presenter presenter.InvoicePresenter) *InvoiceController {
	return &InvoiceController{
		service:   service,
		presenter: presenter,
	}
}

func (c *InvoiceController) CreateManualInvoice(ctx context.Context, req *subscriptionsv1.CreateManualInvoiceRequest) (*subscriptionsv1.InvoiceResponse, error) {
	userID := uuidutil.ParseOrHash(req.GetUserId())
	if userID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	subID, _ := uuid.Parse(req.GetSubscriptionId())

	var dueDate time.Time
	if req.GetDueDate() != "" {
		dueDate, _ = time.Parse(time.RFC3339, req.GetDueDate())
	}

	items := make([]service.CreateInvoiceItemInput, 0, len(req.GetItems()))
	for _, itm := range req.GetItems() {
		items = append(items, service.CreateInvoiceItemInput{
			Description: itm.GetDescription(),
			Quantity:    itm.GetQuantity(),
			UnitPrice:   itm.GetUnitPrice(),
		})
	}

	invoice, err := c.service.CreateManualInvoice(ctx, service.CreateManualInvoiceInput{
		UserID:                  userID,
		SubscriptionID:          subID,
		Amount:                  req.GetAmount(),
		Tax:                     req.GetTax(),
		Currency:                req.GetCurrency(),
		DueDate:                 dueDate,
		Notes:                   req.GetNotes(),
		BankAccountInstructions: req.GetBankAccountInstructions(),
		Items:                   items,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return c.presenter.ToProto(invoice), nil
}

func (c *InvoiceController) GetInvoice(ctx context.Context, req *subscriptionsv1.GetInvoiceRequest) (*subscriptionsv1.InvoiceResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid invoice ID")
	}
	invoice, err := c.service.GetInvoice(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return c.presenter.ToProto(invoice), nil
}

func (c *InvoiceController) ListInvoices(ctx context.Context, req *subscriptionsv1.ListInvoicesRequest) (*subscriptionsv1.ListInvoicesResponse, error) {
	var userIDPtr *uuid.UUID
	if req.GetUserId() != "" {
		uid := uuidutil.ParseOrHash(req.GetUserId())
		if uid != uuid.Nil {
			userIDPtr = &uid
		}
	}

	invoices, total, err := c.service.ListInvoices(ctx, int(req.GetPage()), int(req.GetPageSize()), userIDPtr, req.GetStatus(), req.GetSearch())
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

	return c.presenter.ToListProto(invoices, total, page, pageSize), nil
}

func (c *InvoiceController) MarkInvoicePaid(ctx context.Context, req *subscriptionsv1.MarkInvoicePaidRequest) (*subscriptionsv1.InvoiceResponse, error) {
	id, err := uuid.Parse(req.GetInvoiceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid invoice ID")
	}

	invoice, err := c.service.MarkInvoicePaid(ctx, id, req.GetPaymentReference(), req.GetPaymentMethod(), req.GetAdminNotes())
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return c.presenter.ToProto(invoice), nil
}

func (c *InvoiceController) VoidInvoice(ctx context.Context, req *subscriptionsv1.VoidInvoiceRequest) (*subscriptionsv1.VoidInvoiceResponse, error) {
	id, err := uuid.Parse(req.GetInvoiceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid invoice ID")
	}

	invoice, err := c.service.VoidInvoice(ctx, id, req.GetReason())
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &subscriptionsv1.VoidInvoiceResponse{
		Success: true,
		Message: "Invoice voided successfully",
		Invoice: c.presenter.ToProto(invoice),
	}, nil
}
