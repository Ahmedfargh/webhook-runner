package repository

import (
	"context"
	"errors"

	"subscriptions/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlanRepository interface {
	Create(ctx context.Context, plan *models.Plan) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Plan, error)
	FindByCode(ctx context.Context, code string) (*models.Plan, error)
	Update(ctx context.Context, plan *models.Plan) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, includeInactive bool) ([]models.Plan, error)
}

type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) Create(ctx context.Context, plan *models.Plan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *planRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Plan, error) {
	var plan models.Plan
	if err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) FindByCode(ctx context.Context, code string) (*models.Plan, error) {
	var plan models.Plan
	if err := r.db.WithContext(ctx).First(&plan, "code = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) Update(ctx context.Context, plan *models.Plan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *planRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Plan{}, "id = ?", id).Error
}

func (r *planRepository) List(ctx context.Context, includeInactive bool) ([]models.Plan, error) {
	var plans []models.Plan
	query := r.db.WithContext(ctx).Model(&models.Plan{}).Order("tier_level ASC")
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}
