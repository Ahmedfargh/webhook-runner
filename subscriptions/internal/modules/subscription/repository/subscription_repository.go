package repository

import (
	"context"
	"errors"

	"subscriptions/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, sub *models.Subscription) error
	List(ctx context.Context, page, pageSize int, status string, search string) ([]models.Subscription, int64, error)
}

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *subscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	if err := r.db.WithContext(ctx).Preload("Plan").First(&sub, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	if err := r.db.WithContext(ctx).Preload("Plan").Order("created_at DESC").First(&sub, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *subscriptionRepository) List(ctx context.Context, page, pageSize int, status string, search string) ([]models.Subscription, int64, error) {
	var subs []models.Subscription
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Subscription{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Plan").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&subs).Error; err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}
