package seeders

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"accounts/internal/config"
	"accounts/internal/helpers/passwords"
	"accounts/internal/models"

	"gorm.io/gorm"
)

// AdminSeedJSON matches the structure in admins.json including a temporary CountryCode field
type AdminSeedJSON struct {
	models.Admin
	CountryCode string `json:"country_code"`
	RawPassword string `json:"password"`
}

func SeedAdminsFromFile(db *gorm.DB, filePath string) error {
	fileData, err := os.ReadFile(config.PROJECT_PATH + "/internal/seeders/" + filePath)
	if err != nil {
		return err
	}

	var adminSeeds []AdminSeedJSON
	if err := json.Unmarshal(fileData, &adminSeeds); err != nil {
		return err
	}

	for _, item := range adminSeeds {
		a := item.Admin

		// 1. Look up the specific Country by its CountryCode from JSON
		var country models.Country
		if err := db.Where("country_code = ?", item.CountryCode).First(&country).Error; err != nil {
			log.Printf("Failed to seed admin %s: Country code '%s' not found in database", a.Email, item.CountryCode)
			continue
		}
		a.CountryID = country.ID
		a.Country = country

		// 2. Check if admin already exists
		var existingAdmin models.Admin
		err := db.Where("email = ?", a.Email).First(&existingAdmin).Error
		if err != nil {
			// Extract and clear associations before initial create
			roles := a.Roles
			perms := a.Permissions
			a.Roles = nil
			a.Permissions = nil

			// 3. Hash the password provided in the JSON payload
			passwordToHash := item.RawPassword
			if passwordToHash == "" {
				passwordToHash = "password" // Fallback default if blank
			}

			hashed_password, err := passwords.HashPassword(passwordToHash)
			if err != nil {
				fmt.Println("error in hashing admin's password with email:" + a.Email)
				continue
			}
			a.Password = hashed_password

			// 4. Create the admin record
			if err := db.Create(&a).Error; err != nil {
				log.Printf("Failed to seed admin %s: %v", a.Email, err)
				continue
			}

			// Assign Roles
			for _, role := range roles {
				var dbRole models.Role
				if err := db.Where("name = ?", role.Name).First(&dbRole).Error; err == nil {
					db.Model(&a).Association("Roles").Append(&dbRole)
				}
			}

			// Assign Direct Permissions
			for _, p := range perms {
				var dbPerm models.Permission
				if err := db.Where("name = ?", p.Name).First(&dbPerm).Error; err == nil {
					db.Model(&a).Association("Permissions").Append(&dbPerm)
				}
			}
		}
	}

	log.Println("Admins seeded successfully from file!")
	return nil
}
