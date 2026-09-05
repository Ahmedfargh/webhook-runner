package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"accounts/internal/models"
	"accounts/internal/seeders"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var PROJECT_PATH string

func ConnectDB() {
	PROJECT_PATH, _ = os.Getwd()
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file, falling back to system environment variables")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	var err error
	maxRetries := 15
	for attempt := 1; attempt <= maxRetries; attempt++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			if sqlDB, err := DB.DB(); err == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					sqlDB.SetMaxOpenConns(200)
					sqlDB.SetMaxIdleConns(100)
					log.Println("Accounts database connection established successfully")

					// 1. Auto-migrate all accounts tables
					DB.Exec("SET FOREIGN_KEY_CHECKS = 0;")
					if err := DB.AutoMigrate(
						&models.Country{},
						&models.Permission{},
						&models.Role{},
						&models.User{},
						&models.Admin{},
					); err != nil {
						DB.Exec("SET FOREIGN_KEY_CHECKS = 1;")
						log.Fatalf("Failed to auto-migrate accounts tables: %v", err)
					}
					DB.Exec("SET FOREIGN_KEY_CHECKS = 1;")
					log.Println("Accounts database tables auto-migrated successfully")

					// 2. Auto-seed initial reference data if tables are empty
					seeders.SeedCountriesFromFile(DB)
					if err := seeders.SeedPermissionsFromFile(DB); err != nil {
						log.Printf("Warning: Permissions seeding: %v\n", err)
					}
					if err := seeders.SeedRolesFromFile(DB, "roles.json"); err != nil {
						log.Printf("Warning: Roles seeding: %v\n", err)
					}
					if err := seeders.SeedAdminsFromFile(DB, "admins.json"); err != nil {
						log.Printf("Warning: Admin seeding: %v\n", err)
					}

					return
				}
			}
		}

		log.Printf("Waiting for database connection at %s:%s (attempt %d/%d): %v\n",
			os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), attempt, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
}
