package repository

import (
	"webhookRunner/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookRepository interface {
	Create(call *models.WebhookCall) error
	FindByID(id uuid.UUID) (*models.WebhookCall, error)
	List(userID, appID uuid.UUID, status string, page, limit int, search string) ([]models.WebhookCall, int64, error)
	Update(call *models.WebhookCall) error
}

type webhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) Create(call *models.WebhookCall) error {
	return r.db.Create(call).Error
}

func (r *webhookRepository) FindByID(id uuid.UUID) (*models.WebhookCall, error) {
	var call models.WebhookCall
	if err := r.db.Preload("App").Where("id = ?", id).First(&call).Error; err != nil {
		return nil, err
	}
	return &call, nil
}

func (r *webhookRepository) List(userID, appID uuid.UUID, status string, page, limit int, search string) ([]models.WebhookCall, int64, error) {
	var calls []models.WebhookCall
	var total int64

	query := r.db.Model(&models.WebhookCall{}).Preload("App")

	if appID != uuid.Nil {
		query = query.Where("webhook_calls.app_id = ?", appID)
	} else if userID != uuid.Nil {
		query = query.Joins("JOIN apps ON apps.id = webhook_calls.app_id").Where("apps.user_id = ?", userID)
	}

	if status != "" {
		query = query.Where("webhook_calls.status = ?", status)
	}

	if search != "" {
		likeQuery := "%" + search + "%"
		query = query.Where("webhook_calls.event_name LIKE ? OR webhook_calls.target_url LIKE ? OR webhook_calls.payload_json LIKE ?", likeQuery, likeQuery, likeQuery)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	if err := query.Order("webhook_calls.created_at DESC").Limit(limit).Offset(offset).Find(&calls).Error; err != nil {
		return nil, 0, err
	}

	return calls, total, nil
}

func (r *webhookRepository) Update(call *models.WebhookCall) error {
	return r.db.Save(call).Error
}
