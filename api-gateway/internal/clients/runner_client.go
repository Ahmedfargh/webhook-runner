package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	runnerv1 "webhookApiGateway/api/proto/runner"
	"webhookApiGateway/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RunnerClient struct {
	conn    *grpc.ClientConn
	App     runnerv1.AppServiceClient
	Webhook runnerv1.WebhookRunnerServiceClient
}

func NewRunnerClient(cfg *config.Config) (*RunnerClient, error) {
	target := fmt.Sprintf("%s:%s", cfg.RunnerGRPCHost, cfg.RunnerGRPCPort)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(serviceAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
		grpc.WithStreamInterceptor(serviceStreamAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for %s: %w", target, err)
	}

	client := &RunnerClient{
		conn:    conn,
		App:     runnerv1.NewAppServiceClient(conn),
		Webhook: runnerv1.NewWebhookRunnerServiceClient(conn),
	}

	log.Printf("Connected to Webhook Runner gRPC service at %s (service: %s)\n", target, cfg.ServiceName)
	return client, nil
}

func (c *RunnerClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *RunnerClient) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := c.App.ListApps(timeoutCtx, &runnerv1.ListAppsRequest{
		Page:  1,
		Limit: 1,
	})
	elapsed := time.Since(start)
	return elapsed, err
}
