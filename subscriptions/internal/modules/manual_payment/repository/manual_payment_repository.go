package repository

import (
	"context"
	"errors"

	"subscriptions/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ManualPaymentRepository interface {
	Create(ctx context.Context, payment *models.ManualPaymentRecord) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.ManualPaymentRecord, error)
	FindByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]models.ManualPaymentRecord, error)
	Update(ctx context.Context, payment *models.ManualPaymentRecord) error
	List(ctx context.Context, page, pageSize int, status string, search string) ([]models.ManualPaymentRecord, int64, error)
}

type manualPaymentRepository struct {
	db *gorm.DB
}

func NewManualPaymentRepository(db *gorm.DB) ManualPaymentRepository {
	return &manualPaymentRepository{db: db}
}

func (r *manualPaymentRepository) Create(ctx context.Context, payment *models.ManualPaymentRecord) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *manualPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.ManualPaymentRecord, error) {
	var record models.ManualPaymentRecord
	if err := r.db.WithContext(ctx).Preload("Invoice").First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *manualPaymentRepository) FindByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]models.ManualPaymentRecord, error) {
	var records []models.ManualPaymentRecord
	if err := r.db.WithContext(ctx).Where("invoice_id = ?", invoiceID).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *manualPaymentRepository) Update(ctx context.Context, payment *models.ManualPaymentRecord) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

func (r *manualPaymentRepository) List(ctx context.Context, page, pageSize int, status string, search string) ([]models.ManualPaymentRecord, int64, error) {
	var records []models.ManualPaymentRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ManualPaymentRecord{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("transaction_reference LIKE ? OR payer_name LIKE ? OR payer_notes LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Invoice").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
