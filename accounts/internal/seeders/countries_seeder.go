package seeders

import (
	"encoding/json"
	"fmt"
	"os"

	"accounts/internal/models"

	"gorm.io/gorm"
)

func SeedCountriesFromFile(db *gorm.DB) error {
	fileData, err := os.ReadFile("internal/seeders/countries.json")
	if err != nil {
		fileData = CountriesJSON
	}
	var countries []models.Country
	if err := json.Unmarshal(fileData, &countries); err != nil {
		return err
	}
	for _, country := range countries {
		var existing models.Country
		err := db.Where("country_code = ?", country.CountryCode).First(&existing).Error
		if err != nil {
			if err := db.Create(&country).Error; err != nil {
				fmt.Printf("Failed to seed country %s: %v\n", country.CountryCode, err)
			}
		}
	}
	fmt.Println("Countries seeded successfully!")
	return nil
}
