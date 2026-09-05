package repository

import (
	"webhookRunner/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppRepository interface {
	Create(app *models.App) error
	FindByID(id uuid.UUID) (*models.App, error)
	FindByAppID(appID string) (*models.App, error)
	ListByUserID(userID uuid.UUID, page, limit int, search string) ([]models.App, int64, error)
	Update(app *models.App) error
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type appRepository struct {
	db *gorm.DB
}

func NewAppRepository(db *gorm.DB) AppRepository {
	return &appRepository{db: db}
}

func (r *appRepository) Create(app *models.App) error {
	return r.db.Create(app).Error
}

func (r *appRepository) FindByID(id uuid.UUID) (*models.App, error) {
	var app models.App
	if err := r.db.Where("id = ?", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appRepository) FindByAppID(appID string) (*models.App, error) {
	var app models.App
	if err := r.db.Where("app_id = ?", appID).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appRepository) ListByUserID(userID uuid.UUID, page, limit int, search string) ([]models.App, int64, error) {
	var apps []models.App
	var total int64

	query := r.db.Model(&models.App{})
	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	}

	if search != "" {
		likeQuery := "%" + search + "%"
		query = query.Where("name LIKE ? OR app_id LIKE ? OR webhook_url LIKE ?", likeQuery, likeQuery, likeQuery)
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

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

func (r *appRepository) Update(app *models.App) error {
	return r.db.Save(app).Error
}

func (r *appRepository) Delete(id uuid.UUID, userID uuid.UUID) error {
	query := r.db.Where("id = ?", id)
	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	}
	return query.Delete(&models.App{}).Error
}
