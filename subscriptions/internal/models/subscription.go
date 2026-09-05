package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionStatus string

const (
	StatusActive              SubscriptionStatus = "active"
	StatusPendingManualPayment SubscriptionStatus = "pending_manual_payment"
	StatusTrialing            SubscriptionStatus = "trialing"
	StatusPastDue             SubscriptionStatus = "past_due"
	StatusCanceled            SubscriptionStatus = "canceled"
	StatusExpired             SubscriptionStatus = "expired"
)

type BillingCycle string

const (
	BillingCycleMonthly  BillingCycle = "monthly"
	BillingCycleAnnually BillingCycle = "annually"
)

type Subscription struct {
	ID                 uuid.UUID          `gorm:"type:char(36);primaryKey;" json:"id"`
	UserID             uuid.UUID          `gorm:"type:char(36);index;not null" json:"user_id"`
	PlanID             uuid.UUID          `gorm:"type:char(36);not null" json:"plan_id"`
	Plan               Plan               `gorm:"foreignKey:PlanID" json:"plan"`
	Status             SubscriptionStatus `gorm:"type:varchar(50);not null;default:'pending_manual_payment'" json:"status"`
	BillingCycle       BillingCycle       `gorm:"type:varchar(20);not null;default:'monthly'" json:"billing_cycle"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	TrialEndsAt        *time.Time         `json:"trial_ends_at,omitempty"`
	CancelAtPeriodEnd  bool               `gorm:"default:false" json:"cancel_at_period_end"`
	CustomNotes        string             `gorm:"type:text" json:"custom_notes"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	DeletedAt          gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
