package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"subscriptions/internal/models"
	"subscriptions/internal/seeders"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file in subscriptions service")
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
					log.Println("Subscriptions database connection established successfully")

					// Auto-migrate tables
					DB.Exec("SET FOREIGN_KEY_CHECKS = 0;")
					if err := DB.AutoMigrate(
						&models.Plan{},
						&models.Subscription{},
						&models.Invoice{},
						&models.InvoiceItem{},
						&models.ManualPaymentRecord{},
					); err != nil {
						DB.Exec("SET FOREIGN_KEY_CHECKS = 1;")
						log.Fatal("Failed to auto-migrate subscriptions database:", err)
					}
					DB.Exec("SET FOREIGN_KEY_CHECKS = 1;")
					log.Println("Subscriptions database tables auto-migrated successfully")

					// Auto-seed default subscription plans
					seeders.SeedPlans(DB)

					return
				}
			}
		}

		log.Printf("Waiting for subscriptions database at %s:%s (attempt %d/%d): %v\n",
			os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), attempt, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("Failed to connect to subscriptions database after %d attempts: %v", maxRetries, err)
}
