package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"auditService/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Config struct {
	GRPCPort        string
	AuthToken       string
	AllowedServices []string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	KafkaBrokers    string
	KafkaTopicAudit string
	KafkaGroupID    string
	KafkaEnabled    bool
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or could not be loaded in audit service, using system environment variables")
	}

	port := getEnv("GRPC_PORT", "50054")
	authToken := getEnv("AUTH_TOKEN", "4f7f956f34bcfa0c9a55aff6b98c4e1d87e1da6d0d33f5021b5937123d7330c1")

	allowedServicesRaw := getEnv("ALLOWED_SERVICES", "api-gateway,accounts,subscriptions,webhook-runner,audit-service")
	var allowedServices []string
	if allowedServicesRaw != "" {
		for _, s := range strings.Split(allowedServicesRaw, ",") {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				allowedServices = append(allowedServices, trimmed)
			}
		}
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "webhook_audit")

	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaTopicAudit := getEnv("KAFKA_TOPIC_AUDIT_EVENTS", "audit-events")
	kafkaGroupID := getEnv("KAFKA_GROUP_ID", "audit-service-group")
	kafkaEnabled := getEnv("KAFKA_ENABLED", "true") == "true" || getEnv("KAFKA_ENABLED", "1") == "1"

	return &Config{
		GRPCPort:        port,
		AuthToken:       authToken,
		AllowedServices: allowedServices,
		DBHost:          dbHost,
		DBPort:          dbPort,
		DBUser:          dbUser,
		DBPassword:      dbPassword,
		DBName:          dbName,
		KafkaBrokers:    kafkaBrokers,
		KafkaTopicAudit: kafkaTopicAudit,
		KafkaGroupID:    kafkaGroupID,
		KafkaEnabled:    kafkaEnabled,
	}
}

func ConnectDB(optionalCfg ...*Config) (*gorm.DB, error) {
	var cfg *Config
	if len(optionalCfg) > 0 && optionalCfg[0] != nil {
		cfg = optionalCfg[0]
	} else {
		cfg = LoadConfig()
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
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
					log.Println("Audit database connection established successfully")

					// Auto-migrate tables
					DB.Exec("SET FOREIGN_KEY_CHECKS = 0;")
					if err := DB.AutoMigrate(&models.AuditLog{}); err != nil {
						DB.Exec("SET FOREIGN_KEY_CHECKS = 1;")
						return nil, fmt.Errorf("failed to auto-migrate audit database: %w", err)
					}
					DB.Exec("SET FOREIGN_KEY_CHECKS = 1;")
					log.Println("Audit database tables auto-migrated successfully")

					return DB, nil
				}
			}
		}

		log.Printf("Waiting for audit database at %s:%s (attempt %d/%d): %v\n",
			cfg.DBHost, cfg.DBPort, attempt, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to audit database after %d attempts: %w", maxRetries, err)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
