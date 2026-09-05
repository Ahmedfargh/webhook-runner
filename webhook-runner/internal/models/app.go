package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type App struct {
	ID            uuid.UUID      `gorm:"type:char(36);primaryKey;" json:"id"`
	UserID        uuid.UUID      `gorm:"type:char(36);index;not null" json:"user_id"`
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	AppID         string         `gorm:"type:varchar(60);uniqueIndex;not null" json:"app_id"`
	AppSecret     string         `gorm:"type:varchar(120);not null" json:"app_secret"`
	WebhookURL    string         `gorm:"type:varchar(500);not null" json:"webhook_url"`
	WebhookSecret string         `gorm:"type:varchar(100);not null" json:"webhook_secret"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *App) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
