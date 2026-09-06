package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"webhookRunner/internal/config"
	"webhookRunner/internal/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	cfg        *config.Config
	container  *Container
	grpcServer *grpc.Server
}

func New(cfg *config.Config) (*App, error) {
	db, err := config.ConnectDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}

	container := NewContainer(db, cfg)

	authInterceptor := middleware.NewAuthInterceptor(cfg.AuthToken, cfg.AllowedServices)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	container.RegisterGRPCServices(grpcServer)
	reflection.Register(grpcServer)

	return &App{
		cfg:        cfg,
		container:  container,
		grpcServer: grpcServer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	address := fmt.Sprintf(":%s", a.cfg.GRPCPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to bind port %s: %w", address, err)
	}

	// 1. Start background Kafka Consumer Worker
	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()

	if a.container.Consumer != nil && a.container.Consumer.IsEnabled() {
		go a.container.Consumer.Start(consumerCtx)
	}

	// 2. Start gRPC Server
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Webhook Runner gRPC server listening on %s (Protected)\n", address)
		if err := a.grpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Context cancelled. Initiating graceful shutdown...")
	case sig := <-stopChan:
		log.Printf("Received signal %v. Initiating graceful shutdown...\n", sig)
	case err := <-serverErr:
		return fmt.Errorf("gRPC server encountered an unexpected error: %w", err)
	}

	// 3. Graceful Shutdown Coordinator
	cancelConsumer()
	a.container.Close()
	a.grpcServer.GracefulStop()

	log.Println("Webhook Runner gRPC server and Kafka worker stopped cleanly.")
	return nil
}
