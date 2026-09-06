package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "webhookApiGateway/api/proto/request_tracker/v1"
	"webhookApiGateway/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestTrackerClient struct {
	conn    *grpc.ClientConn
	Tracker pb.RequestTrackerServiceClient
}

func NewRequestTrackerClient(cfg *config.Config) (*RequestTrackerClient, error) {
	addr := fmt.Sprintf("%s:%s", cfg.RequestTrackerGRPCHost, cfg.RequestTrackerGRPCPort)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(serviceAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
		grpc.WithStreamInterceptor(serviceStreamAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		log.Printf("Warning: Failed to create gRPC client for Request Tracker service at %s: %v", addr, err)
		return &RequestTrackerClient{}, nil
	}

	return &RequestTrackerClient{
		conn:    conn,
		Tracker: pb.NewRequestTrackerServiceClient(conn),
	}, nil
}

func (c *RequestTrackerClient) Ping(ctx context.Context) (time.Duration, error) {
	if c.Tracker == nil {
		return 0, fmt.Errorf("request tracker gRPC client not initialized")
	}
	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := c.Tracker.ListTraces(pingCtx, &pb.ListTracesRequest{Limit: 1})
	return time.Since(start), err
}

func (c *RequestTrackerClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
