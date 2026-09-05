package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"subscriptions/internal/models"
	"subscriptions/internal/modules/plan/repository"
	"subscriptions/internal/pkg/apperrors"

	"github.com/google/uuid"
)

type CreatePlanInput struct {
	Name              string
	Code              string
	Description       string
	PriceMonthly      float64
	PriceAnnually     float64
	Currency          string
	MaxWebhooks       int32
	MaxEventsPerMonth int64
	MaxTeamMembers    int32
	Features          []string
	IsActive          bool
	IsPopular         bool
	TierLevel         int32
}

type UpdatePlanInput struct {
	ID                uuid.UUID
	Name              string
	Description       string
	PriceMonthly      float64
	PriceAnnually     float64
	Currency          string
	MaxWebhooks       int32
	MaxEventsPerMonth int64
	MaxTeamMembers    int32
	Features          []string
	IsActive          bool
	IsPopular         bool
	TierLevel         int32
}

type PlanService interface {
	CreatePlan(ctx context.Context, input CreatePlanInput) (*models.Plan, error)
	GetPlan(ctx context.Context, id uuid.UUID) (*models.Plan, error)
	UpdatePlan(ctx context.Context, input UpdatePlanInput) (*models.Plan, error)
	DeletePlan(ctx context.Context, id uuid.UUID) error
	ListPlans(ctx context.Context, includeInactive bool) ([]models.Plan, error)
}

type planService struct {
	repo repository.PlanRepository
}

func NewPlanService(repo repository.PlanRepository) PlanService {
	return &planService{repo: repo}
}

func (s *planService) CreatePlan(ctx context.Context, input CreatePlanInput) (*models.Plan, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: plan name cannot be empty", apperrors.ErrInvalidArgument)
	}

	code := strings.ToLower(strings.TrimSpace(input.Code))
	if code == "" {
		code = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}

	existing, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: plan code '%s' already exists", apperrors.ErrAlreadyExists, code)
	}

	featuresJSON, _ := json.Marshal(input.Features)

	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = "USD"
	}

	plan := &models.Plan{
		ID:                uuid.New(),
		Name:              name,
		Code:              code,
		Description:       input.Description,
		PriceMonthly:      input.PriceMonthly,
		PriceAnnually:     input.PriceAnnually,
		Currency:          currency,
		MaxWebhooks:       input.MaxWebhooks,
		MaxEventsPerMonth: input.MaxEventsPerMonth,
		MaxTeamMembers:    input.MaxTeamMembers,
		FeaturesJSON:      string(featuresJSON),
		IsActive:          input.IsActive,
		IsPopular:         input.IsPopular,
		TierLevel:         input.TierLevel,
	}

	if err := s.repo.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	return plan, nil
}

func (s *planService) GetPlan(ctx context.Context, id uuid.UUID) (*models.Plan, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid plan ID", apperrors.ErrInvalidArgument)
	}
	plan, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find plan: %w", err)
	}
	if plan == nil {
		return nil, fmt.Errorf("%w: plan not found", apperrors.ErrNotFound)
	}
	return plan, nil
}

func (s *planService) UpdatePlan(ctx context.Context, input UpdatePlanInput) (*models.Plan, error) {
	if input.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid plan ID", apperrors.ErrInvalidArgument)
	}

	plan, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, fmt.Errorf("%w: plan not found", apperrors.ErrNotFound)
	}

	if input.Name != "" {
		plan.Name = strings.TrimSpace(input.Name)
	}
	if input.Description != "" {
		plan.Description = input.Description
	}
	plan.PriceMonthly = input.PriceMonthly
	plan.PriceAnnually = input.PriceAnnually
	if input.Currency != "" {
		plan.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	}
	plan.MaxWebhooks = input.MaxWebhooks
	plan.MaxEventsPerMonth = input.MaxEventsPerMonth
	plan.MaxTeamMembers = input.MaxTeamMembers
	if input.Features != nil {
		featuresJSON, _ := json.Marshal(input.Features)
		plan.FeaturesJSON = string(featuresJSON)
	}
	plan.IsActive = input.IsActive
	plan.IsPopular = input.IsPopular
	plan.TierLevel = input.TierLevel

	if err := s.repo.Update(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	return plan, nil
}

func (s *planService) DeletePlan(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: invalid plan ID", apperrors.ErrInvalidArgument)
	}
	plan, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if plan == nil {
		return fmt.Errorf("%w: plan not found", apperrors.ErrNotFound)
	}
	return s.repo.Delete(ctx, id)
}

func (s *planService) ListPlans(ctx context.Context, includeInactive bool) ([]models.Plan, error) {
	return s.repo.List(ctx, includeInactive)
}
