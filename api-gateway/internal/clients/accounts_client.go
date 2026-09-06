package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	accountsv1 "webhookApiGateway/api/proto/v1"
	"webhookApiGateway/internal/config"
	"webhookApiGateway/internal/telemetry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type AccountsClient struct {
	conn       *grpc.ClientConn
	User       accountsv1.UserServiceClient
	Admin      accountsv1.AdminServiceClient
	Role       accountsv1.RoleServiceClient
	Permission accountsv1.PermissionServiceClient
	Country    accountsv1.CountryServiceClient
}

// serviceAuthInterceptor adds x-service-name, x-trace-id and authorization metadata headers and records the gRPC trip
func serviceAuthInterceptor(serviceName, serviceToken string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		collector := telemetry.GetSpanCollector(ctx)

		traceID := ""
		if collector != nil {
			traceID = collector.GetTraceID()
		}

		pairs := []string{
			"x-service-name", serviceName,
			"authorization", "Bearer " + serviceToken,
		}
		if traceID != "" {
			pairs = append(pairs, "x-trace-id", traceID, "x-request-id", traceID)
		}

		md := metadata.Pairs(pairs...)
		ctx = metadata.NewOutgoingContext(ctx, md)

		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		if collector != nil {
			targetService := "accounts-service"
			if cc != nil {
				targetService = cc.Target()
			}
			status := "OK"
			if err != nil {
				status = "ERROR"
			}
			collector.AddSpan("gRPC: "+method, targetService, "GRPC", "DOWNSTREAM_RPC", start, duration, status, "")
		}

		return err
	}
}

// serviceStreamAuthInterceptor adds metadata headers to all outgoing stream gRPC calls
func serviceStreamAuthInterceptor(serviceName, serviceToken string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		collector := telemetry.GetSpanCollector(ctx)
		traceID := ""
		if collector != nil {
			traceID = collector.GetTraceID()
		}

		pairs := []string{
			"x-service-name", serviceName,
			"authorization", "Bearer " + serviceToken,
		}
		if traceID != "" {
			pairs = append(pairs, "x-trace-id", traceID, "x-request-id", traceID)
		}

		md := metadata.Pairs(pairs...)
		ctx = metadata.NewOutgoingContext(ctx, md)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func NewAccountsClient(cfg *config.Config) (*AccountsClient, error) {
	target := fmt.Sprintf("%s:%s", cfg.AccountsGRPCHost, cfg.AccountsGRPCPort)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(serviceAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
		grpc.WithStreamInterceptor(serviceStreamAuthInterceptor(cfg.ServiceName, cfg.ServiceToken)),
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for %s: %w", target, err)
	}

	client := &AccountsClient{
		conn:       conn,
		User:       accountsv1.NewUserServiceClient(conn),
		Admin:      accountsv1.NewAdminServiceClient(conn),
		Role:       accountsv1.NewRoleServiceClient(conn),
		Permission: accountsv1.NewPermissionServiceClient(conn),
		Country:    accountsv1.NewCountryServiceClient(conn),
	}

	log.Printf("Connected to Accounts gRPC service at %s (service: %s)\n", target, cfg.ServiceName)
	return client, nil
}

func (c *AccountsClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Ping checks if the accounts service is reachable by querying permissions or users
func (c *AccountsClient) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := c.Permission.ListPermissions(timeoutCtx, &accountsv1.ListPermissionsRequest{
		Pagination: &accountsv1.PaginationRequest{Page: 1, PageSize: 1},
	})
	elapsed := time.Since(start)
	return elapsed, err
}
