package models

import (
	"accounts/internal/helpers/phonenumbers"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Admin struct {
	ID          uuid.UUID    `gorm:"type:char(36);primaryKey;" json:"id"`
	Name        string       `gorm:"type:varchar(191);not null" json:"name"`
	Email       string       `gorm:"type:varchar(191);uniqueIndex;not null" json:"email"`
	Phone       string       `gorm:"type:varchar(20);not null" json:"phone"`
	Password    string       `gorm:"type:varchar(255);not null" json:"-"`
	CountryID   uuid.UUID    `gorm:"type:char(36);not null" json:"country_id"`
	Country     Country      `gorm:"foreignKey:CountryID" json:"country"`
	Roles       []Role       `gorm:"many2many:admin_has_role;constraint:OnDelete:CASCADE;" json:"roles"`
	Permissions []Permission `gorm:"many2many:admin_has_permission;constraint:OnDelete:CASCADE;" json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (u *Admin) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	if u.Country.CountryCode == "" && u.CountryID != uuid.Nil {
		var country Country
		if err := tx.First(&country, "id = ?", u.CountryID).Error; err == nil {
			u.Country = country
		}
	}

	u.Phone, err = phonenumbers.NormalizePhoneNumber(u.Phone, u.Country.CountryCode)
	if err != nil {
		return err
	}
	return nil
}
