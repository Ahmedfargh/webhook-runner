package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	AccountsGRPCHost      string
	AccountsGRPCPort      string
	SubscriptionsGRPCHost string
	SubscriptionsGRPCPort string
	RunnerGRPCHost        string
	RunnerGRPCPort        string
	AuditGRPCHost         string
	AuditGRPCPort         string
	RequestTrackerGRPCHost string
	RequestTrackerGRPCPort string
	ServiceName           string
	ServiceToken          string
	JWTSecret             string
	AllowedOrigins        string
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	KafkaBrokers          string
	KafkaTopicDispatch    string
	KafkaTopicResults     string
	KafkaTopicAudit       string
	KafkaTopicRequestTraces string
	KafkaEnabled          bool
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or could not be loaded, using system environment variables")
	}

	port := getEnv("PORT", "8080")
	accountsHost := getEnv("ACCOUNTS_GRPC_HOST", "localhost")
	accountsPort := getEnv("ACCOUNTS_GRPC_PORT", "50051")
	subscriptionsHost := getEnv("SUBSCRIPTIONS_GRPC_HOST", "localhost")
	subscriptionsPort := getEnv("SUBSCRIPTIONS_GRPC_PORT", "50052")
	runnerHost := getEnv("RUNNER_GRPC_HOST", "localhost")
	runnerPort := getEnv("RUNNER_GRPC_PORT", "50053")
	auditHost := getEnv("AUDIT_GRPC_HOST", "localhost")
	auditPort := getEnv("AUDIT_GRPC_PORT", "50054")
	requestTrackerHost := getEnv("REQUEST_TRACKER_GRPC_HOST", "localhost")
	requestTrackerPort := getEnv("REQUEST_TRACKER_GRPC_PORT", "50055")
	serviceName := getEnv("SERVICE_NAME", "api-gateway")
	serviceToken := getEnv("SERVICE_TOKEN", "4f7f956f34bcfa0c9a55aff6b98c4e1d87e1da6d0d33f5021b5937123d7330c1")
	jwtSecret := getEnv("JWT_SECRET", "api-gateway-super-secret-jwt-key-2026")
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "*")

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "webhook_accounts")

	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaTopicDispatch := getEnv("KAFKA_TOPIC_WEBHOOK_DISPATCH", "webhook-dispatches")
	kafkaTopicResults := getEnv("KAFKA_TOPIC_WEBHOOK_RESULTS", "webhook-results")
	kafkaTopicAudit := getEnv("KAFKA_TOPIC_AUDIT_EVENTS", "audit-events")
	kafkaTopicRequestTraces := getEnv("KAFKA_TOPIC_REQUEST_TRACES", "http-request-traces")
	kafkaEnabled := getEnv("KAFKA_ENABLED", "true") == "true" || getEnv("KAFKA_ENABLED", "1") == "1"

	return &Config{
		Port:                   port,
		AccountsGRPCHost:       accountsHost,
		AccountsGRPCPort:       accountsPort,
		SubscriptionsGRPCHost:  subscriptionsHost,
		SubscriptionsGRPCPort:  subscriptionsPort,
		RunnerGRPCHost:         runnerHost,
		RunnerGRPCPort:         runnerPort,
		AuditGRPCHost:          auditHost,
		AuditGRPCPort:          auditPort,
		RequestTrackerGRPCHost: requestTrackerHost,
		RequestTrackerGRPCPort: requestTrackerPort,
		ServiceName:            serviceName,
		ServiceToken:           serviceToken,
		JWTSecret:              jwtSecret,
		AllowedOrigins:         allowedOrigins,
		DBHost:                 dbHost,
		DBPort:                 dbPort,
		DBUser:                 dbUser,
		DBPassword:             dbPassword,
		DBName:                 dbName,
		KafkaBrokers:           kafkaBrokers,
		KafkaTopicDispatch:     kafkaTopicDispatch,
		KafkaTopicResults:      kafkaTopicResults,
		KafkaTopicAudit:        kafkaTopicAudit,
		KafkaTopicRequestTraces: kafkaTopicRequestTraces,
		KafkaEnabled:           kafkaEnabled,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
