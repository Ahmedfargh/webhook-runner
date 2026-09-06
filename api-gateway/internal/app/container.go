package app

import (
	"fmt"

	"webhookApiGateway/internal/audit"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/config"
	"webhookApiGateway/internal/handlers"
	"webhookApiGateway/internal/kafka"
)

type Container struct {
	AccountsClient      *clients.AccountsClient
	SubscriptionsClient *clients.SubscriptionsClient
	RunnerClient        *clients.RunnerClient
	AuditClient         *clients.AuditClient
	KafkaProducer       *kafka.KafkaProducer
	AuditEmitter        *audit.KafkaEmitter

	AuthHandler          *handlers.AuthHandler
	UserHandler          *handlers.UserHandler
	AdminHandler         *handlers.AdminHandler
	RoleHandler          *handlers.RoleHandler
	PermHandler          *handlers.PermissionHandler
	HealthHandler        *handlers.HealthHandler
	CountryHandler       *handlers.CountryHandler
	PlanHandler          *handlers.PlanHandler
	SubscriptionHandler  *handlers.SubscriptionHandler
	InvoiceHandler       *handlers.InvoiceHandler
	ManualPaymentHandler *handlers.ManualPaymentHandler
	AppHandler           *handlers.AppHandler
	WebhookHandler       *handlers.WebhookHandler
	AuditHandler         *handlers.AuditHandler
}

func NewContainer(cfg *config.Config) (*Container, error) {
	accountsClient, err := clients.NewAccountsClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Accounts gRPC client: %w", err)
	}

	subscriptionsClient, err := clients.NewSubscriptionsClient(cfg)
	if err != nil {
		accountsClient.Close()
		return nil, fmt.Errorf("failed to initialize Subscriptions gRPC client: %w", err)
	}

	runnerClient, err := clients.NewRunnerClient(cfg)
	if err != nil {
		accountsClient.Close()
		subscriptionsClient.Close()
		return nil, fmt.Errorf("failed to initialize Runner gRPC client: %w", err)
	}

	auditClient, err := clients.NewAuditClient(cfg)
	if err != nil {
		accountsClient.Close()
		subscriptionsClient.Close()
		runnerClient.Close()
		return nil, fmt.Errorf("failed to initialize Audit gRPC client: %w", err)
	}

	kafkaProducer := kafka.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaTopicDispatch, cfg.KafkaEnabled)
	auditEmitter := audit.NewEmitter(cfg.KafkaBrokers, cfg.KafkaTopicAudit, "api-gateway", cfg.KafkaEnabled)

	authHandler := handlers.NewAuthHandler(accountsClient, cfg)
	userHandler := handlers.NewUserHandler(accountsClient)
	adminHandler := handlers.NewAdminHandler(accountsClient)
	roleHandler := handlers.NewRoleHandler(accountsClient)
	permHandler := handlers.NewPermissionHandler(accountsClient)
	healthHandler := handlers.NewHealthHandler(accountsClient, subscriptionsClient, runnerClient, kafkaProducer, cfg)
	countryHandler := handlers.NewCountryHandler(accountsClient)

	planHandler := handlers.NewPlanHandler(subscriptionsClient)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionsClient)
	invoiceHandler := handlers.NewInvoiceHandler(subscriptionsClient)
	manualPaymentHandler := handlers.NewManualPaymentHandler(subscriptionsClient)

	appHandler := handlers.NewAppHandler(runnerClient)
	webhookHandler := handlers.NewWebhookHandler(runnerClient, kafkaProducer)
	auditHandler := handlers.NewAuditHandler(auditClient)

	return &Container{
		AccountsClient:       accountsClient,
		SubscriptionsClient:  subscriptionsClient,
		RunnerClient:         runnerClient,
		AuditClient:          auditClient,
		KafkaProducer:        kafkaProducer,
		AuditEmitter:         auditEmitter,
		AuthHandler:          authHandler,
		UserHandler:          userHandler,
		AdminHandler:         adminHandler,
		RoleHandler:          roleHandler,
		PermHandler:          permHandler,
		HealthHandler:        healthHandler,
		CountryHandler:       countryHandler,
		PlanHandler:          planHandler,
		SubscriptionHandler:  subscriptionHandler,
		InvoiceHandler:       invoiceHandler,
		ManualPaymentHandler: manualPaymentHandler,
		AppHandler:           appHandler,
		WebhookHandler:       webhookHandler,
		AuditHandler:         auditHandler,
	}, nil
}

func (c *Container) Close() {
	if c.AccountsClient != nil {
		_ = c.AccountsClient.Close()
	}
	if c.SubscriptionsClient != nil {
		_ = c.SubscriptionsClient.Close()
	}
	if c.RunnerClient != nil {
		_ = c.RunnerClient.Close()
	}
	if c.AuditClient != nil {
		_ = c.AuditClient.Close()
	}
	if c.KafkaProducer != nil {
		_ = c.KafkaProducer.Close()
	}
	if c.AuditEmitter != nil {
		_ = c.AuditEmitter.Close()
	}
}
