package models

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	Admins    []Admin   `gorm:"many2many:admin_has_permission;" json:"admins,omitempty"`
	Roles     []Role    `gorm:"many2many:role_has_permission;" json:"roles,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
