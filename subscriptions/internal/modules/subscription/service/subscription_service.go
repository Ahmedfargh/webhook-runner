package service

import (
	"context"
	"fmt"
	"time"

	"subscriptions/internal/models"
	invoiceService "subscriptions/internal/modules/invoice/service"
	planRepo "subscriptions/internal/modules/plan/repository"
	"subscriptions/internal/modules/subscription/repository"
	"subscriptions/internal/pkg/apperrors"

	"github.com/google/uuid"
)

type SubscribeInput struct {
	UserID        uuid.UUID
	PlanID        uuid.UUID
	BillingCycle  models.BillingCycle
	PaymentMethod string
	Notes         string
}

type SubscribeResult struct {
	Subscription        *models.Subscription
	InvoiceID           string
	InvoiceNumber       string
	AmountDue           float64
	Currency            string
	PaymentInstructions string
}

type AdminOverrideInput struct {
	UserID           uuid.UUID
	PlanID           uuid.UUID
	Status           models.SubscriptionStatus
	CurrentPeriodEnd *time.Time
	AdminNotes       string
}

type SubscriptionService interface {
	Subscribe(ctx context.Context, input SubscribeInput) (*SubscribeResult, error)
	GetUserSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error)
	CancelSubscription(ctx context.Context, userID uuid.UUID, reason string, immediately bool) (*models.Subscription, error)
	AdminOverrideSubscription(ctx context.Context, input AdminOverrideInput) (*models.Subscription, error)
	ListSubscriptions(ctx context.Context, page, pageSize int, status string, search string) ([]models.Subscription, int64, error)
}

type subscriptionService struct {
	repo       repository.SubscriptionRepository
	planRepo   planRepo.PlanRepository
	invoiceSvc invoiceService.InvoiceService
}

func NewSubscriptionService(
	repo repository.SubscriptionRepository,
	planRepo planRepo.PlanRepository,
	invoiceSvc invoiceService.InvoiceService,
) SubscriptionService {
	return &subscriptionService{
		repo:       repo,
		planRepo:   planRepo,
		invoiceSvc: invoiceSvc,
	}
}

func (s *subscriptionService) Subscribe(ctx context.Context, input SubscribeInput) (*SubscribeResult, error) {
	if input.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid user ID", apperrors.ErrInvalidArgument)
	}
	if input.PlanID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid plan ID", apperrors.ErrInvalidArgument)
	}

	plan, err := s.planRepo.FindByID(ctx, input.PlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, fmt.Errorf("%w: plan not found", apperrors.ErrNotFound)
	}

	now := time.Now()
	var periodEnd time.Time
	if input.BillingCycle == models.BillingCycleAnnually {
		periodEnd = now.AddDate(1, 0, 0)
	} else {
		input.BillingCycle = models.BillingCycleMonthly
		periodEnd = now.AddDate(0, 1, 0)
	}

	existingSub, _ := s.repo.FindByUserID(ctx, input.UserID)

	// Is Free Plan?
	isFree := plan.PriceMonthly == 0 && plan.PriceAnnually == 0
	subStatus := models.StatusPendingManualPayment
	if isFree {
		subStatus = models.StatusActive
		periodEnd = now.AddDate(100, 0, 0)
	}

	var sub *models.Subscription
	if existingSub != nil {
		sub = existingSub
		sub.PlanID = plan.ID
		sub.Plan = *plan
		sub.Status = subStatus
		sub.BillingCycle = input.BillingCycle
		sub.CurrentPeriodStart = now
		sub.CurrentPeriodEnd = periodEnd
		sub.CancelAtPeriodEnd = false
		sub.CustomNotes = input.Notes
		if err := s.repo.Update(ctx, sub); err != nil {
			return nil, fmt.Errorf("failed to update subscription: %w", err)
		}
	} else {
		sub = &models.Subscription{
			ID:                 uuid.New(),
			UserID:             input.UserID,
			PlanID:             plan.ID,
			Plan:               *plan,
			Status:             subStatus,
			BillingCycle:       input.BillingCycle,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   periodEnd,
			CancelAtPeriodEnd:  false,
			CustomNotes:        input.Notes,
		}
		if err := s.repo.Create(ctx, sub); err != nil {
			return nil, fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	res := &SubscribeResult{
		Subscription: sub,
	}

	// If paid plan, generate unpaid invoice with offline bank wire instructions
	if !isFree {
		inv, err := s.invoiceSvc.CreateSubscriptionInvoice(ctx, sub, plan, input.BillingCycle)
		if err != nil {
			return nil, fmt.Errorf("failed to generate subscription invoice: %w", err)
		}
		res.InvoiceID = inv.ID.String()
		res.InvoiceNumber = inv.InvoiceNumber
		res.AmountDue = inv.TotalAmount
		res.Currency = inv.Currency
		res.PaymentInstructions = inv.BankAccountInstructions
	}

	return res, nil
}

func (s *subscriptionService) GetUserSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid user ID", apperrors.ErrInvalidArgument)
	}

	sub, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if sub != nil {
		return sub, nil
	}

	// Auto-provision Free plan
	freePlan, err := s.planRepo.FindByCode(ctx, "free")
	if err != nil || freePlan == nil {
		plans, _ := s.planRepo.List(ctx, false)
		if len(plans) > 0 {
			freePlan = &plans[0]
		}
	}

	if freePlan == nil {
		return nil, fmt.Errorf("%w: no default plan available", apperrors.ErrNotFound)
	}

	now := time.Now()
	newSub := &models.Subscription{
		ID:                 uuid.New(),
		UserID:             userID,
		PlanID:             freePlan.ID,
		Plan:               *freePlan,
		Status:             models.StatusActive,
		BillingCycle:       models.BillingCycleMonthly,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(100, 0, 0),
		CancelAtPeriodEnd:  false,
	}

	if err := s.repo.Create(ctx, newSub); err != nil {
		return nil, err
	}

	return newSub, nil
}

func (s *subscriptionService) CancelSubscription(ctx context.Context, userID uuid.UUID, reason string, immediately bool) (*models.Subscription, error) {
	sub, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("%w: subscription not found", apperrors.ErrNotFound)
	}

	if immediately {
		sub.Status = models.StatusCanceled
	} else {
		sub.CancelAtPeriodEnd = true
	}
	if reason != "" {
		sub.CustomNotes += "\n[Cancellation]: " + reason
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) AdminOverrideSubscription(ctx context.Context, input AdminOverrideInput) (*models.Subscription, error) {
	if input.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid user ID", apperrors.ErrInvalidArgument)
	}

	sub, err := s.repo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("%w: subscription not found", apperrors.ErrNotFound)
	}

	if input.PlanID != uuid.Nil {
		plan, err := s.planRepo.FindByID(ctx, input.PlanID)
		if err != nil {
			return nil, err
		}
		if plan != nil {
			sub.PlanID = plan.ID
			sub.Plan = *plan
		}
	}

	if input.Status != "" {
		sub.Status = input.Status
	}
	if input.CurrentPeriodEnd != nil {
		sub.CurrentPeriodEnd = *input.CurrentPeriodEnd
	}
	if input.AdminNotes != "" {
		sub.CustomNotes += "\n[Admin Override]: " + input.AdminNotes
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, page, pageSize int, status string, search string) ([]models.Subscription, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.List(ctx, page, pageSize, status, search)
}
