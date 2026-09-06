package app

import (
	pb "requestTrackerService/api/proto/v1"
	"requestTrackerService/internal/config"
	"requestTrackerService/internal/kafka"
	"requestTrackerService/internal/modules/trace/presenter"
	"requestTrackerService/internal/repository"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Container struct {
	DB          *gorm.DB
	TraceRepo   repository.TraceRepository
	GRPCHandler *presenter.GRPCHandler
	Consumer    *kafka.Consumer
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	traceRepo := repository.NewTraceRepository(db)
	grpcHandler := presenter.NewGRPCHandler(traceRepo)

	consumer := kafka.NewConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopicRequestTraces,
		cfg.KafkaGroupID,
		cfg.KafkaEnabled,
		traceRepo,
	)

	return &Container{
		DB:          db,
		TraceRepo:   traceRepo,
		GRPCHandler: grpcHandler,
		Consumer:    consumer,
	}
}

func (c *Container) RegisterGRPCServices(server *grpc.Server) {
	pb.RegisterRequestTrackerServiceServer(server, c.GRPCHandler)
}

func (c *Container) Close() {
	if c.Consumer != nil {
		_ = c.Consumer.Close()
	}
}
