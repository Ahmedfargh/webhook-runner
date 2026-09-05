package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusSubmitted PaymentStatus = "submitted"
	PaymentStatusApproved  PaymentStatus = "approved"
	PaymentStatusRejected  PaymentStatus = "rejected"
)

type ManualPaymentRecord struct {
	ID                   uuid.UUID      `gorm:"type:char(36);primaryKey;" json:"id"`
	InvoiceID            uuid.UUID      `gorm:"type:char(36);index;not null" json:"invoice_id"`
	Invoice              Invoice        `gorm:"foreignKey:InvoiceID" json:"invoice"`
	UserID               uuid.UUID      `gorm:"type:char(36);index;not null" json:"user_id"`
	Amount               float64        `gorm:"type:decimal(10,2);not null;default:0.00" json:"amount"`
	Currency             string         `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"`
	PaymentMethod        string         `gorm:"type:varchar(50);not null;default:'bank_transfer'" json:"payment_method"`
	TransactionReference string         `gorm:"type:varchar(100);not null" json:"transaction_reference"`
	PayerName            string         `gorm:"type:varchar(100)" json:"payer_name"`
	PayerNotes           string         `gorm:"type:text" json:"payer_notes"`
	RecordedByAdminID    *uuid.UUID     `gorm:"type:char(36)" json:"recorded_by_admin_id,omitempty"`
	AdminNotes           string         `gorm:"type:text" json:"admin_notes"`
	Status               PaymentStatus  `gorm:"type:varchar(20);not null;default:'submitted'" json:"status"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *ManualPaymentRecord) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
