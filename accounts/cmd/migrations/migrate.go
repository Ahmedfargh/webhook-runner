package main

import (
	"log"

	"accounts/internal/config"
	"accounts/internal/models"
)

func main() {
	// Initialize the database connection
	config.ConnectDB()

	// Run auto-migrations with error handling
	err := config.DB.AutoMigrate(
		&models.Admin{},
		&models.User{},
		&models.Country{},
		&models.Role{},
		&models.Permission{},
	)
	if err != nil {
		log.Fatalf("Failed to run auto-migrations: %v", err)
	}

	log.Println("Database migration completed successfully!")
}
