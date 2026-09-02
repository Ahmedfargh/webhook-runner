package models

import (
	"accounts/internal/helpers/phonenumbers"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Phone     string    `gorm:"type:varchar(20);not null" json:"phone"`
	Password  string    `gorm:"not null" json:"-"`
	CountryID uuid.UUID `gorm:"type:uuid;not null" json:"country_id"`
	Country   Country   `gorm:"foreignKey:CountryID" json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.Phone, err = phonenumbers.NormalizePhoneNumber(u.Phone, u.Country.CountryCode)
	if err != nil {
		return err
	}
	return nil
}
