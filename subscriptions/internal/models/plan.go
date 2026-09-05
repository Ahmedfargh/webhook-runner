package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Plan struct {
	ID                uuid.UUID      `gorm:"type:char(36);primaryKey;" json:"id"`
	Name              string         `gorm:"type:varchar(100);not null" json:"name"`
	Code              string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Description       string         `gorm:"type:text" json:"description"`
	PriceMonthly      float64        `gorm:"type:decimal(10,2);not null;default:0.00" json:"price_monthly"`
	PriceAnnually     float64        `gorm:"type:decimal(10,2);not null;default:0.00" json:"price_annually"`
	Currency          string         `gorm:"type:varchar(10);not null;default:'USD'" json:"currency"`
	MaxWebhooks       int32          `gorm:"type:int;not null;default:5" json:"max_webhooks"`
	MaxEventsPerMonth int64          `gorm:"type:bigint;not null;default:10000" json:"max_events_per_month"`
	MaxTeamMembers    int32          `gorm:"type:int;not null;default:1" json:"max_team_members"`
	FeaturesJSON      string         `gorm:"type:text" json:"features_json"` // JSON encoded string array
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	IsPopular         bool           `gorm:"default:false" json:"is_popular"`
	TierLevel         int32          `gorm:"default:1" json:"tier_level"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Plan) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
