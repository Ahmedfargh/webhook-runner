package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookCallStatus string

const (
	StatusPending  WebhookCallStatus = "pending"
	StatusSuccess  WebhookCallStatus = "success"
	StatusFailed   WebhookCallStatus = "failed"
	StatusRetrying WebhookCallStatus = "retrying"
)

type WebhookCall struct {
	ID                 uuid.UUID         `gorm:"type:char(36);primaryKey;" json:"id"`
	AppID              uuid.UUID         `gorm:"type:char(36);index;not null" json:"app_id"`
	App                App               `gorm:"foreignKey:AppID;references:ID;-:migration" json:"app"`
	EventName          string            `gorm:"type:varchar(100);index;not null" json:"event_name"`
	TargetURL          string            `gorm:"type:varchar(500);not null" json:"target_url"`
	PayloadJSON        string            `gorm:"type:longtext;not null" json:"payload_json"`
	HeadersJSON        string            `gorm:"type:text" json:"headers_json"`
	AttemptCount       int32             `gorm:"default:1" json:"attempt_count"`
	Status             WebhookCallStatus `gorm:"type:varchar(50);index;not null;default:'pending'" json:"status"`
	ResponseStatusCode int32             `gorm:"default:0" json:"response_status_code"`
	ResponseBody       string            `gorm:"type:text" json:"response_body"`
	ResponseLatencyMS  int64             `gorm:"default:0" json:"response_latency_ms"`
	ErrorMessage       string            `gorm:"type:text" json:"error_message"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	DeletedAt          gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (w *WebhookCall) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
