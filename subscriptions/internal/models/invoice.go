package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft   InvoiceStatus = "draft"
	InvoiceStatusUnpaid  InvoiceStatus = "unpaid"
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusVoid    InvoiceStatus = "void"
	InvoiceStatusOverdue InvoiceStatus = "overdue"
)

type Invoice struct {
	ID                      uuid.UUID      `gorm:"type:char(36);primaryKey;" json:"id"`
	InvoiceNumber           string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"invoice_number"`
	SubscriptionID          uuid.UUID      `gorm:"type:char(36);index" json:"subscription_id"`
	UserID                  uuid.UUID      `gorm:"type:char(36);index;not null" json:"user_id"`
	Amount                  float64        `gorm:"type:decimal(10,2);not null;default:0.00" json:"amount"`
	Tax                     float64        `gorm:"type:decimal(10,2);not null;default:0.00" json:"tax"`
	TotalAmount             float64        `gorm:"type:decimal(10,2);not null;default:0.00" json:"total_amount"`
	Currency                string         `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"`
	Status                  InvoiceStatus  `gorm:"type:varchar(20);not null;default:'unpaid'" json:"status"`
	DueDate                 time.Time      `json:"due_date"`
	PaidAt                  *time.Time     `json:"paid_at,omitempty"`
	PaymentMethod           string         `gorm:"type:varchar(50);default:'bank_transfer'" json:"payment_method"`
	PaymentReference        string         `gorm:"type:varchar(100)" json:"payment_reference"`
	BankAccountInstructions string         `gorm:"type:text" json:"bank_account_instructions"`
	Notes                   string         `gorm:"type:text" json:"notes"`
	Items                   []InvoiceItem  `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE;" json:"items"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`
}

type InvoiceItem struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey;" json:"id"`
	InvoiceID   uuid.UUID      `gorm:"type:char(36);index;not null" json:"invoice_id"`
	Description string         `gorm:"type:varchar(255);not null" json:"description"`
	Quantity    int32          `gorm:"type:int;not null;default:1" json:"quantity"`
	UnitPrice   float64        `gorm:"type:decimal(10,2);not null;default:0.00" json:"unit_price"`
	Total       float64        `gorm:"type:decimal(10,2);not null;default:0.00" json:"total"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	if i.TotalAmount == 0 {
		i.TotalAmount = i.Amount + i.Tax
	}
	return nil
}

func (item *InvoiceItem) BeforeCreate(tx *gorm.DB) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if item.Total == 0 {
		item.Total = float64(item.Quantity) * item.UnitPrice
	}
	return nil
}
