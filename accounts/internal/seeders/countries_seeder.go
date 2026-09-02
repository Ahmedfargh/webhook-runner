package seeders

import (
	"encoding/json"
	"fmt"
	"os"

	"accounts/internal/config"
	"accounts/internal/models"

	"gorm.io/gorm"
)

func SeedCountriesFromFile(db *gorm.DB) error {

	fileData, err := os.ReadFile(config.PROJECT_PATH + "/internal/seeders/countries.json")
	if err != nil {
		fmt.Println(err)
		return err
	}
	var countries []models.Country
	if err := json.Unmarshal(fileData, &countries); err != nil {
		return err
	}
	for _, country := range countries {
		var existing models.Country
		err := config.DB.Where("country_code = ?", country.CountryCode).First(&existing).Error
		if err != nil {
			// Record not found, create a new one
			if err := config.DB.Create(&country).Error; err != nil {
				fmt.Printf("Failed to seed country %s: %v", country.CountryCode, err)
			}
		}
	}
	fmt.Println("Countries seeded successfully from file!")
	return nil
}
