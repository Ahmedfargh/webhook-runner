package main

import (
	"accounts/internal/config"
	"accounts/internal/seeders"
	"log"

	"gorm.io/gorm"
)

func RunSeeders(db *gorm.DB) {
	seeders.SeedCountriesFromFile(config.DB)
	if err := seeders.SeedPermissionsFromFile(db); err != nil {
		log.Fatalf("Permission seeding failed: %v", err)
	}

	if err := seeders.SeedRolesFromFile(db, "roles.json"); err != nil {
		log.Fatalf("Role seeding failed: %v", err)
	}

	if err := seeders.SeedAdminsFromFile(db, "admins.json"); err != nil {
		log.Fatalf("Admin seeding failed: %v", err)
	}
}
func main() {
	config.ConnectDB()
	RunSeeders(config.DB)
}
