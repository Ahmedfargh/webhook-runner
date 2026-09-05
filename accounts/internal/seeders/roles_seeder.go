package seeders

import (
	"encoding/json"
	"log"
	"os"

	"accounts/internal/models"

	"gorm.io/gorm"
)

func SeedRolesFromFile(db *gorm.DB, filePath string) error {
	fileData, err := os.ReadFile("internal/seeders/" + filePath)
	if err != nil {
		fileData = RolesJSON
	}

	var roles []models.Role
	if err := json.Unmarshal(fileData, &roles); err != nil {
		return err
	}

	for _, r := range roles {
		var existingRole models.Role
		err := db.Where("name = ?", r.Name).First(&existingRole).Error
		if err != nil {
			// Extract and clear nested permissions to avoid duplicate key issues during creation
			perms := r.Permissions
			r.Permissions = nil

			if err := db.Create(&r).Error; err != nil {
				log.Printf("Failed to seed role %s: %v", r.Name, err)
				continue
			}

			// Attach valid permissions via GORM associations
			for _, p := range perms {
				var dbPerm models.Permission
				if err := db.Where("name = ?", p.Name).First(&dbPerm).Error; err == nil {
					db.Model(&r).Association("Permissions").Append(&dbPerm)
				}
			}
		}
	}

	log.Println("Roles seeded successfully!")
	return nil
}
