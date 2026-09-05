package seeders

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"accounts/internal/helpers/passwords"
	"accounts/internal/models"

	"gorm.io/gorm"
)

type UserSeedJSON struct {
	models.User
	CountryCode string `json:"country_code"`
	RawPassword string `json:"password"`
}

func SeedUsersFromFile(db *gorm.DB, filePath string) error {
	fileData, err := os.ReadFile("internal/seeders/" + filePath)
	if err != nil {
		fileData = UsersJSON
	}

	var userSeeds []UserSeedJSON
	if err := json.Unmarshal(fileData, &userSeeds); err != nil {
		return err
	}

	for _, item := range userSeeds {
		u := item.User

		// 1. Look up Country by CountryCode
		var country models.Country
		if err := db.Where("country_code = ?", item.CountryCode).First(&country).Error; err != nil {
			log.Printf("Failed to seed user %s: Country code '%s' not found", u.Email, item.CountryCode)
			continue
		}
		u.CountryID = country.ID
		u.Country = country

		// 2. Check if user exists
		var existingUser models.User
		if err := db.Where("email = ?", u.Email).First(&existingUser).Error; err != nil {
			passwordToHash := item.RawPassword
			if passwordToHash == "" {
				passwordToHash = "password123"
			}

			hashedPassword, err := passwords.HashPassword(passwordToHash)
			if err != nil {
				fmt.Println("error hashing password for user: " + u.Email)
				continue
			}
			u.Password = hashedPassword

			if err := db.Create(&u).Error; err != nil {
				log.Printf("Failed to seed user %s: %v", u.Email, err)
				continue
			}
		}
	}

	log.Println("Users seeded successfully!")
	return nil
}
