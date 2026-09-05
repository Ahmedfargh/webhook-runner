package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"subscriptions/internal/models"
	invoiceService "subscriptions/internal/modules/invoice/service"
	"subscriptions/internal/modules/manual_payment/repository"
	subRepo "subscriptions/internal/modules/subscription/repository"
	"subscriptions/internal/pkg/apperrors"

	"github.com/google/uuid"
)

type SubmitPaymentInput struct {
	InvoiceID            uuid.UUID
	UserID               uuid.UUID
	Amount               float64
	Currency             string
	PaymentMethod        string
	TransactionReference string
	PayerName            string
	PayerNotes           string
}

type ReviewPaymentInput struct {
	PaymentID  uuid.UUID
	Approve    bool
	AdminNotes string
	AdminID    *uuid.UUID
}

type ManualPaymentService interface {
	SubmitPayment(ctx context.Context, input SubmitPaymentInput) (*models.ManualPaymentRecord, error)
	ReviewPayment(ctx context.Context, input ReviewPaymentInput) (*models.ManualPaymentRecord, error)
	ListPayments(ctx context.Context, page, pageSize int, status string, search string) ([]models.ManualPaymentRecord, int64, error)
}

type manualPaymentService struct {
	repo       repository.ManualPaymentRepository
	invoiceSvc invoiceService.InvoiceService
	subRepo    subRepo.SubscriptionRepository
}

func NewManualPaymentService(
	repo repository.ManualPaymentRepository,
	invoiceSvc invoiceService.InvoiceService,
	subRepo subRepo.SubscriptionRepository,
) ManualPaymentService {
	return &manualPaymentService{
		repo:       repo,
		invoiceSvc: invoiceSvc,
		subRepo:    subRepo,
	}
}

func (s *manualPaymentService) SubmitPayment(ctx context.Context, input SubmitPaymentInput) (*models.ManualPaymentRecord, error) {
	if input.InvoiceID == uuid.Nil {
		return nil, fmt.Errorf("%w: invoice ID is required", apperrors.ErrInvalidArgument)
	}

	invoice, err := s.invoiceSvc.GetInvoice(ctx, input.InvoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil {
		return nil, fmt.Errorf("%w: invoice not found", apperrors.ErrNotFound)
	}

	if invoice.Status == models.InvoiceStatusPaid {
		return nil, fmt.Errorf("%w: this invoice has already been paid", apperrors.ErrInvoiceAlreadyPaid)
	}

	amount := input.Amount
	if amount <= 0 {
		amount = invoice.TotalAmount
	}

	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = invoice.Currency
	}

	paymentMethod := strings.TrimSpace(input.PaymentMethod)
	if paymentMethod == "" {
		paymentMethod = "bank_transfer"
	}

	record := &models.ManualPaymentRecord{
		ID:                   uuid.New(),
		InvoiceID:            input.InvoiceID,
		UserID:               input.UserID,
		Amount:               amount,
		Currency:             currency,
		PaymentMethod:        paymentMethod,
		TransactionReference: strings.TrimSpace(input.TransactionReference),
		PayerName:            strings.TrimSpace(input.PayerName),
		PayerNotes:           strings.TrimSpace(input.PayerNotes),
		Status:               models.PaymentStatusSubmitted,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to record manual payment: %w", err)
	}

	return record, nil
}

func (s *manualPaymentService) ReviewPayment(ctx context.Context, input ReviewPaymentInput) (*models.ManualPaymentRecord, error) {
	if input.PaymentID == uuid.Nil {
		return nil, fmt.Errorf("%w: payment ID is required", apperrors.ErrInvalidArgument)
	}

	payment, err := s.repo.FindByID(ctx, input.PaymentID)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, fmt.Errorf("%w: payment record not found", apperrors.ErrNotFound)
	}

	payment.AdminNotes = strings.TrimSpace(input.AdminNotes)
	payment.RecordedByAdminID = input.AdminID

	if input.Approve {
		payment.Status = models.PaymentStatusApproved

		// 1. Mark invoice as paid
		invoice, err := s.invoiceSvc.MarkInvoicePaid(ctx, payment.InvoiceID, payment.TransactionReference, payment.PaymentMethod, input.AdminNotes)
		if err != nil {
			return nil, fmt.Errorf("failed to mark invoice paid: %w", err)
		}

		// 2. If subscription linked, activate subscription & extend period
		if invoice.SubscriptionID != uuid.Nil {
			sub, err := s.subRepo.FindByID(ctx, invoice.SubscriptionID)
			if err == nil && sub != nil {
				now := time.Now()
				sub.Status = models.StatusActive
				sub.CurrentPeriodStart = now
				if sub.BillingCycle == models.BillingCycleAnnually {
					sub.CurrentPeriodEnd = now.AddDate(1, 0, 0)
				} else {
					sub.CurrentPeriodEnd = now.AddDate(0, 1, 0)
				}
				_ = s.subRepo.Update(ctx, sub)
			}
		}
	} else {
		payment.Status = models.PaymentStatusRejected
	}

	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment review: %w", err)
	}

	return payment, nil
}

func (s *manualPaymentService) ListPayments(ctx context.Context, page, pageSize int, status string, search string) ([]models.ManualPaymentRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.List(ctx, page, pageSize, strings.TrimSpace(status), strings.TrimSpace(search))
}
