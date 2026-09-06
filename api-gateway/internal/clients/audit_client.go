package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "webhookApiGateway/api/proto/audit/v1"
	"webhookApiGateway/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuditClient struct {
	conn  *grpc.ClientConn
	Audit pb.AuditServiceClient
}

func NewAuditClient(cfg *config.Config) (*AuditClient, error) {
	addr := fmt.Sprintf("%s:%s", cfg.AuditGRPCHost, cfg.AuditGRPCPort)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(serviceAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
		grpc.WithStreamInterceptor(serviceStreamAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		log.Printf("Warning: Failed to create gRPC client for Audit service at %s: %v", addr, err)
		return &AuditClient{}, nil
	}

	return &AuditClient{
		conn:  conn,
		Audit: pb.NewAuditServiceClient(conn),
	}, nil
}

func (c *AuditClient) Ping(ctx context.Context) (time.Duration, error) {
	if c.Audit == nil {
		return 0, fmt.Errorf("audit gRPC client not initialized")
	}
	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := c.Audit.ListAuditLogs(pingCtx, &pb.ListAuditLogsRequest{Limit: 1})
	return time.Since(start), err
}

func (c *AuditClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
