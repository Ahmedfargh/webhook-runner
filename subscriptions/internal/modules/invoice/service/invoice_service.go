package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"subscriptions/internal/models"
	"subscriptions/internal/modules/invoice/repository"
	"subscriptions/internal/pkg/apperrors"

	"github.com/google/uuid"
)

const DefaultBankInstructions = `=== Offline Bank Transfer Instructions ===
Bank Name: Standard International Bank
Account Name: Webhook Cloud Services Inc.
Account Number: 8820-9931-4402-1198
IBAN: US92 SIBK 8820 9931 4402 1198
SWIFT/BIC: SIBKUS33XXX
Routing / Transit: 021000021

Payment Reference Required: Please include your Invoice Number in the transfer memo.
After sending payment, submit your wire/transaction reference in the billing portal for instant verification.`

type CreateInvoiceItemInput struct {
	Description string
	Quantity    int32
	UnitPrice   float64
}

type CreateManualInvoiceInput struct {
	UserID                  uuid.UUID
	SubscriptionID          uuid.UUID
	Amount                  float64
	Tax                     float64
	Currency                string
	DueDate                 time.Time
	Notes                   string
	BankAccountInstructions string
	Items                   []CreateInvoiceItemInput
}

type InvoiceService interface {
	CreateManualInvoice(ctx context.Context, input CreateManualInvoiceInput) (*models.Invoice, error)
	CreateSubscriptionInvoice(ctx context.Context, sub *models.Subscription, plan *models.Plan, cycle models.BillingCycle) (*models.Invoice, error)
	GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
	ListInvoices(ctx context.Context, page, pageSize int, userID *uuid.UUID, status string, search string) ([]models.Invoice, int64, error)
	MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID, paymentRef string, paymentMethod string, adminNotes string) (*models.Invoice, error)
	VoidInvoice(ctx context.Context, invoiceID uuid.UUID, reason string) (*models.Invoice, error)
}

type invoiceService struct {
	repo repository.InvoiceRepository
}

func NewInvoiceService(repo repository.InvoiceRepository) InvoiceService {
	return &invoiceService{repo: repo}
}

func (s *invoiceService) CreateManualInvoice(ctx context.Context, input CreateManualInvoiceInput) (*models.Invoice, error) {
	if input.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: user ID cannot be empty", apperrors.ErrInvalidArgument)
	}

	invNum, err := s.repo.GenerateNextInvoiceNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = "USD"
	}

	dueDate := input.DueDate
	if dueDate.IsZero() {
		dueDate = time.Now().AddDate(0, 0, 14) // 14 days default
	}

	bankInstructions := input.BankAccountInstructions
	if bankInstructions == "" {
		bankInstructions = DefaultBankInstructions
	}

	items := make([]models.InvoiceItem, 0, len(input.Items))
	totalAmount := 0.0
	for _, itm := range input.Items {
		itemTotal := float64(itm.Quantity) * itm.UnitPrice
		totalAmount += itemTotal
		items = append(items, models.InvoiceItem{
			ID:          uuid.New(),
			Description: itm.Description,
			Quantity:    itm.Quantity,
			UnitPrice:   itm.UnitPrice,
			Total:       itemTotal,
		})
	}

	if totalAmount == 0 && input.Amount > 0 {
		totalAmount = input.Amount
	}

	invoice := &models.Invoice{
		ID:                      uuid.New(),
		InvoiceNumber:           invNum,
		SubscriptionID:          input.SubscriptionID,
		UserID:                  input.UserID,
		Amount:                  totalAmount,
		Tax:                     input.Tax,
		TotalAmount:             totalAmount + input.Tax,
		Currency:                currency,
		Status:                  models.InvoiceStatusUnpaid,
		DueDate:                 dueDate,
		PaymentMethod:           "bank_transfer",
		BankAccountInstructions: bankInstructions,
		Notes:                   input.Notes,
		Items:                   items,
	}

	if err := s.repo.Create(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

func (s *invoiceService) CreateSubscriptionInvoice(ctx context.Context, sub *models.Subscription, plan *models.Plan, cycle models.BillingCycle) (*models.Invoice, error) {
	invNum, err := s.repo.GenerateNextInvoiceNumber(ctx)
	if err != nil {
		return nil, err
	}

	var price float64
	var cycleText string
	if cycle == models.BillingCycleAnnually {
		price = plan.PriceAnnually
		cycleText = "Annual Subscription"
	} else {
		price = plan.PriceMonthly
		cycleText = "Monthly Subscription"
	}

	desc := fmt.Sprintf("%s - %s Plan (%s)", plan.Name, cycleText, plan.Currency)

	items := []models.InvoiceItem{
		{
			ID:          uuid.New(),
			Description: desc,
			Quantity:    1,
			UnitPrice:   price,
			Total:       price,
		},
	}

	invoice := &models.Invoice{
		ID:                      uuid.New(),
		InvoiceNumber:           invNum,
		SubscriptionID:          sub.ID,
		UserID:                  sub.UserID,
		Amount:                  price,
		Tax:                     0.00,
		TotalAmount:             price,
		Currency:                plan.Currency,
		Status:                  models.InvoiceStatusUnpaid,
		DueDate:                 time.Now().AddDate(0, 0, 7), // 7 days to complete bank wire
		PaymentMethod:           "bank_transfer",
		BankAccountInstructions: DefaultBankInstructions,
		Notes:                   fmt.Sprintf("Subscription renewal order for %s plan.", plan.Name),
		Items:                   items,
	}

	if err := s.repo.Create(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to create subscription invoice: %w", err)
	}

	return invoice, nil
}

func (s *invoiceService) GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid invoice ID", apperrors.ErrInvalidArgument)
	}
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if invoice == nil {
		return nil, fmt.Errorf("%w: invoice not found", apperrors.ErrNotFound)
	}
	return invoice, nil
}

func (s *invoiceService) ListInvoices(ctx context.Context, page, pageSize int, userID *uuid.UUID, status string, search string) ([]models.Invoice, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.List(ctx, page, pageSize, userID, strings.TrimSpace(status), strings.TrimSpace(search))
}

func (s *invoiceService) MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID, paymentRef string, paymentMethod string, adminNotes string) (*models.Invoice, error) {
	invoice, err := s.repo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil {
		return nil, fmt.Errorf("%w: invoice not found", apperrors.ErrNotFound)
	}

	if invoice.Status == models.InvoiceStatusPaid {
		return invoice, nil
	}

	now := time.Now()
	invoice.Status = models.InvoiceStatusPaid
	invoice.PaidAt = &now
	if paymentRef != "" {
		invoice.PaymentReference = strings.TrimSpace(paymentRef)
	}
	if paymentMethod != "" {
		invoice.PaymentMethod = strings.TrimSpace(paymentMethod)
	}
	if adminNotes != "" {
		if invoice.Notes != "" {
			invoice.Notes += "\n[Admin]: " + adminNotes
		} else {
			invoice.Notes = "[Admin]: " + adminNotes
		}
	}

	if err := s.repo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	return invoice, nil
}

func (s *invoiceService) VoidInvoice(ctx context.Context, invoiceID uuid.UUID, reason string) (*models.Invoice, error) {
	invoice, err := s.repo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil {
		return nil, fmt.Errorf("%w: invoice not found", apperrors.ErrNotFound)
	}

	if invoice.Status == models.InvoiceStatusPaid {
		return nil, fmt.Errorf("%w: cannot void a paid invoice", apperrors.ErrInvalidArgument)
	}

	invoice.Status = models.InvoiceStatusVoid
	if reason != "" {
		invoice.Notes += "\n[Void Reason]: " + reason
	}

	if err := s.repo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to void invoice: %w", err)
	}

	return invoice, nil
}
