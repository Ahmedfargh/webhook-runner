package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	ActorID      uuid.UUID `gorm:"type:char(36);index;null" json:"actor_id"`
	ActorType    string    `gorm:"type:varchar(50);index;not null;default:'SYSTEM'" json:"actor_type"` // USER, ADMIN, SYSTEM, SERVICE
	ActorName    string    `gorm:"type:varchar(255)" json:"actor_name"`
	ActorEmail   string    `gorm:"type:varchar(191);index" json:"actor_email"`
	ServiceName  string    `gorm:"type:varchar(100);index;not null" json:"service_name"`
	Action       string    `gorm:"type:varchar(100);index;not null" json:"action"`   // CREATE, UPDATE, DELETE, LOGIN, OVERRIDE, ROTATE_SECRET, DISPATCH
	Resource     string    `gorm:"type:varchar(100);index;not null" json:"resource"` // USER, ADMIN, ROLE, PERMISSION, PLAN, SUBSCRIPTION, INVOICE, APP, WEBHOOK
	ResourceID   string    `gorm:"type:varchar(255);index" json:"resource_id"`
	BeforeJSON   string    `gorm:"type:longtext" json:"before_json"`
	AfterJSON    string    `gorm:"type:longtext" json:"after_json"`
	IPAddress    string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	Status       string    `gorm:"type:varchar(50);index;not null;default:'SUCCESS'" json:"status"` // SUCCESS, FAILED
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	return nil
}
