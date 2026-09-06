package repository

import (
	"auditService/internal/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditFilter struct {
	ActorID     string
	ServiceName string
	Resource    string
	Action      string
	Status      string
	Search      string
	StartDate   string
	EndDate     string
	Page        int
	Limit       int
}

type AuditRepository interface {
	Create(log *models.AuditLog) error
	FindByID(id uuid.UUID) (*models.AuditLog, error)
	List(filter AuditFilter) ([]models.AuditLog, int64, error)
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditRepository) FindByID(id uuid.UUID) (*models.AuditLog, error) {
	var log models.AuditLog
	err := r.db.First(&log, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *auditRepository) List(filter AuditFilter) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Model(&models.AuditLog{})

	if filter.ActorID != "" {
		if parsed, err := uuid.Parse(filter.ActorID); err == nil {
			query = query.Where("actor_id = ?", parsed)
		}
	}
	if filter.ServiceName != "" {
		query = query.Where("service_name = ?", filter.ServiceName)
	}
	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.StartDate != "" {
		query = query.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	if filter.Search != "" {
		s := "%" + strings.TrimSpace(filter.Search) + "%"
		query = query.Where("actor_name LIKE ? OR actor_email LIKE ? OR resource_id LIKE ? OR error_message LIKE ?", s, s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
