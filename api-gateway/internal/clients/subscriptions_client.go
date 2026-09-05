package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	subscriptionsv1 "webhookApiGateway/api/proto/subscriptions/v1"
	"webhookApiGateway/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SubscriptionsClient struct {
	conn          *grpc.ClientConn
	Plan          subscriptionsv1.PlanServiceClient
	Subscription  subscriptionsv1.SubscriptionServiceClient
	Invoice       subscriptionsv1.InvoiceServiceClient
	ManualPayment subscriptionsv1.ManualPaymentServiceClient
}

func NewSubscriptionsClient(cfg *config.Config) (*SubscriptionsClient, error) {
	target := fmt.Sprintf("%s:%s", cfg.SubscriptionsGRPCHost, cfg.SubscriptionsGRPCPort)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(serviceAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
		grpc.WithStreamInterceptor(serviceStreamAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for %s: %w", target, err)
	}

	client := &SubscriptionsClient{
		conn:          conn,
		Plan:          subscriptionsv1.NewPlanServiceClient(conn),
		Subscription:  subscriptionsv1.NewSubscriptionServiceClient(conn),
		Invoice:       subscriptionsv1.NewInvoiceServiceClient(conn),
		ManualPayment: subscriptionsv1.NewManualPaymentServiceClient(conn),
	}

	log.Printf("Connected to Subscriptions gRPC service at %s (service: %s)\n", target, cfg.ServiceName)
	return client, nil
}

func (c *SubscriptionsClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *SubscriptionsClient) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := c.Plan.ListPlans(timeoutCtx, &subscriptionsv1.ListPlansRequest{
		IncludeInactive: false,
	})
	elapsed := time.Since(start)
	return elapsed, err
}
