package app

import (
	"webhookRunner/internal/audit"
	"webhookRunner/internal/config"
	"webhookRunner/internal/engine"
	"webhookRunner/internal/kafka"
	"webhookRunner/internal/modules/app"
	"webhookRunner/internal/modules/webhook"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Container struct {
	DB             *gorm.DB
	AuditEmitter   *audit.KafkaEmitter
	Dispatcher     *engine.Dispatcher
	AppModule      *app.AppModule
	WebhookModule  *webhook.WebhookModule
	ResultProducer *kafka.ResultProducer
	Consumer       *kafka.Consumer
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	dispatcher := engine.NewDispatcher()
	appMod := app.NewAppModule(db)
	webhookMod := webhook.NewWebhookModule(db, appMod.Repository, dispatcher)

	resultProducer := kafka.NewResultProducer(cfg.KafkaBrokers, cfg.KafkaTopicResults, cfg.KafkaEnabled)
	consumer := kafka.NewConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopicDispatch,
		cfg.KafkaGroupID,
		cfg.KafkaEnabled,
		webhookMod.Service,
		resultProducer,
	)

	auditEmitter := audit.NewEmitter(cfg.KafkaBrokers, "audit-events", "webhook-runner", cfg.KafkaEnabled)

	return &Container{
		DB:             db,
		AuditEmitter:   auditEmitter,
		Dispatcher:     dispatcher,
		AppModule:      appMod,
		WebhookModule:  webhookMod,
		ResultProducer: resultProducer,
		Consumer:       consumer,
	}
}

func (c *Container) RegisterGRPCServices(server *grpc.Server) {
	c.AppModule.RegisterGRPC(server)
	c.WebhookModule.RegisterGRPC(server)
}

func (c *Container) Close() {
	if c.Consumer != nil {
		_ = c.Consumer.Close()
	}
	if c.ResultProducer != nil {
		_ = c.ResultProducer.Close()
	}
	if c.AuditEmitter != nil {
		_ = c.AuditEmitter.Close()
	}
}
