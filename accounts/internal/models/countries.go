package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LocaleName struct {
	AR string `json:"ar"`
	EN string `json:"en"`
}

type Country struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;" json:"id"`
	Name        LocaleName `gorm:"serializer:json;type:json;not null" json:"name"`
	CountryCode string     `gorm:"type:varchar(10);uniqueIndex;not null" json:"country_code"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (u *Country) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	return nil
}
