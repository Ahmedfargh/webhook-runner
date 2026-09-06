package app

import (
	"auditService/internal/config"
	"auditService/internal/kafka"
	"auditService/internal/modules/audit"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Container struct {
	DB          *gorm.DB
	AuditModule *audit.AuditModule
	Consumer    *kafka.Consumer
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	auditMod := audit.NewAuditModule(db)

	consumer := kafka.NewConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopicAudit,
		cfg.KafkaGroupID,
		cfg.KafkaEnabled,
		auditMod.Service,
	)

	return &Container{
		DB:          db,
		AuditModule: auditMod,
		Consumer:    consumer,
	}
}

func (c *Container) RegisterGRPCServices(server *grpc.Server) {
	c.AuditModule.RegisterGRPC(server)
}

func (c *Container) Close() {
	if c.Consumer != nil {
		_ = c.Consumer.Close()
	}
}
