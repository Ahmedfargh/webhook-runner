package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"subscriptions/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *models.Invoice) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
	FindByInvoiceNumber(ctx context.Context, num string) (*models.Invoice, error)
	Update(ctx context.Context, invoice *models.Invoice) error
	List(ctx context.Context, page, pageSize int, userID *uuid.UUID, status string, search string) ([]models.Invoice, int64, error)
	GenerateNextInvoiceNumber(ctx context.Context) (string, error)
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(ctx context.Context, invoice *models.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

func (r *invoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.WithContext(ctx).Preload("Items").First(&invoice, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) FindByInvoiceNumber(ctx context.Context, num string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.WithContext(ctx).Preload("Items").First(&invoice, "invoice_number = ?", num).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) Update(ctx context.Context, invoice *models.Invoice) error {
	return r.db.WithContext(ctx).Save(invoice).Error
}

func (r *invoiceRepository) List(ctx context.Context, page, pageSize int, userID *uuid.UUID, status string, search string) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Invoice{})

	if userID != nil && *userID != uuid.Nil {
		query = query.Where("user_id = ?", *userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("invoice_number LIKE ? OR payment_reference LIKE ? OR notes LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Items").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

func (r *invoiceRepository) GenerateNextInvoiceNumber(ctx context.Context) (string, error) {
	year := time.Now().Year()
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Invoice{}).Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("INV-%d-%05d", year, count+1), nil
}
