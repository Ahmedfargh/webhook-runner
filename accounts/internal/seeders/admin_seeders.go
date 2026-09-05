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

type AdminSeedJSON struct {
	models.Admin
	CountryCode string `json:"country_code"`
	RawPassword string `json:"password"`
}

func SeedAdminsFromFile(db *gorm.DB, filePath string) error {
	fileData, err := os.ReadFile("internal/seeders/" + filePath)
	if err != nil {
		fileData = AdminsJSON
	}

	var adminSeeds []AdminSeedJSON
	if err := json.Unmarshal(fileData, &adminSeeds); err != nil {
		return err
	}

	for _, item := range adminSeeds {
		a := item.Admin

		// 1. Look up Country by CountryCode
		var country models.Country
		if err := db.Where("country_code = ?", item.CountryCode).First(&country).Error; err != nil {
			log.Printf("Failed to seed admin %s: Country code '%s' not found in database", a.Email, item.CountryCode)
			continue
		}
		a.CountryID = country.ID
		a.Country = country

		// 2. Check if admin exists
		var existingAdmin models.Admin
		err := db.Where("email = ?", a.Email).First(&existingAdmin).Error
		if err != nil {
			roles := a.Roles
			perms := a.Permissions
			a.Roles = nil
			a.Permissions = nil

			passwordToHash := item.RawPassword
			if passwordToHash == "" {
				passwordToHash = "password"
			}

			hashed_password, err := passwords.HashPassword(passwordToHash)
			if err != nil {
				fmt.Println("error in hashing admin password: " + a.Email)
				continue
			}
			a.Password = hashed_password

			if err := db.Create(&a).Error; err != nil {
				log.Printf("Failed to seed admin %s: %v", a.Email, err)
				continue
			}

			for _, role := range roles {
				var dbRole models.Role
				if err := db.Where("name = ?", role.Name).First(&dbRole).Error; err == nil {
					db.Model(&a).Association("Roles").Append(&dbRole)
				}
			}

			for _, p := range perms {
				var dbPerm models.Permission
				if err := db.Where("name = ?", p.Name).First(&dbPerm).Error; err == nil {
					db.Model(&a).Association("Permissions").Append(&dbPerm)
				}
			}
		}
	}

	log.Println("Admins seeded successfully!")
	return nil
}
