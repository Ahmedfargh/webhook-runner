package seeders

import (
	"encoding/json"
	"fmt"
	"os"

	"accounts/internal/config"
	"accounts/internal/models"

	"gorm.io/gorm"
)

func SeedPermissionsFromFile(db *gorm.DB) error {
	fileData, err := os.ReadFile(config.PROJECT_PATH + "/internal/seeders/permissions.json")
	if err != nil {
		fmt.Println(err)
		return err
	}

	var permissions []models.Permission
	if err := json.Unmarshal(fileData, &permissions); err != nil {
		return err
	}

	for _, p := range permissions {
		var existing models.Permission
		if err := config.DB.Where("name = ?", p.Name).First(&existing).Error; err != nil {
			if err := config.DB.Create(&p).Error; err != nil {
				fmt.Printf("Failed to seed permission %s: %v\n", p.Name, err)
			}
		}
	}
	fmt.Println("Permissions seeded successfully from file!")
	return nil
}
