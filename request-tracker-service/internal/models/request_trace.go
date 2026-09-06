package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RequestTrace struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	TraceID      string    `gorm:"type:varchar(64);index;not null" json:"trace_id"`
	RequestID    string    `gorm:"type:varchar(64);index" json:"request_id"`
	ActorType    string    `gorm:"type:varchar(32);index;not null;default:'ANONYMOUS'" json:"actor_type"` // ADMIN, USER, ANONYMOUS, SYSTEM
	ActorID      string    `gorm:"type:varchar(64);index" json:"actor_id"`
	ActorName    string    `gorm:"type:varchar(255)" json:"actor_name"`
	ActorEmail   string    `gorm:"type:varchar(191);index" json:"actor_email"`
	ActorRole    string    `gorm:"type:varchar(64)" json:"actor_role"`
	ServiceName  string    `gorm:"type:varchar(64);index;not null;default:'api-gateway'" json:"service_name"`
	Method       string    `gorm:"type:varchar(16);index;not null" json:"method"`
	Path         string    `gorm:"type:varchar(500);not null" json:"path"`
	Route        string    `gorm:"type:varchar(255);index;not null" json:"route"`
	QueryParams  string    `gorm:"type:text" json:"query_params"`
	ClientIP     string    `gorm:"type:varchar(45);index" json:"client_ip"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	StatusCode   int       `gorm:"type:int;index;not null" json:"status_code"`
	LifetimeMs   float64   `gorm:"type:decimal(12,3);index;not null" json:"lifetime_ms"`
	RequestBody  string    `gorm:"type:longtext" json:"request_body"`
	ResponseBody string    `gorm:"type:longtext" json:"response_body"`
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	SpansJSON    string    `gorm:"type:longtext" json:"spans_json"`
	ReceivedAt   time.Time `gorm:"index;not null" json:"received_at"`
	CompletedAt  time.Time `gorm:"not null" json:"completed_at"`
	CreatedAt    time.Time `gorm:"index;not null" json:"created_at"`
}

func (r *RequestTrace) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	if r.ReceivedAt.IsZero() {
		r.ReceivedAt = time.Now().UTC()
	}
	if r.CompletedAt.IsZero() {
		r.CompletedAt = time.Now().UTC()
	}
	return nil
}
