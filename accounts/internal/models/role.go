package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;" json:"id"`
	Name        string       `gorm:"uniqueIndex;not null" json:"name"`
	Admins      []Admin      `gorm:"many2many:admin_has_role;" json:"admins,omitempty"`
	Permissions []Permission `gorm:"many2many:role_has_permission;" json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
