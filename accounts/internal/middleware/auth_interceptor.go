package middleware

import (
	"context"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	// AuthTokenContextKey stores the authenticated token in context
	AuthTokenContextKey contextKey = "auth_token"
	// ServiceNameContextKey stores the calling service name in context
	ServiceNameContextKey contextKey = "service_name"
)

// AuthInterceptor enforces token-based authentication and service authorization on incoming gRPC requests
type AuthInterceptor struct {
	expectedToken   string
	allowedServices map[string]bool
	exemptMethods   map[string]bool
}

// NewAuthInterceptor creates a new token-based gRPC auth & service authorization interceptor
func NewAuthInterceptor(expectedToken string, allowedServices []string, exemptMethods ...string) *AuthInterceptor {
	if expectedToken == "" {
		expectedToken = os.Getenv("AUTH_TOKEN")
		if expectedToken == "" {
			expectedToken = "webhook-accounts-secret-token"
		}
	}

	allowed := make(map[string]bool)
	if len(allowedServices) == 0 {
		envAllowed := os.Getenv("ALLOWED_SERVICES")
		if envAllowed == "" {
			envAllowed = "api-gateway,webhook-runner"
		}
		for _, s := range strings.Split(envAllowed, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				allowed[s] = true
			}
		}
	} else {
		for _, s := range allowedServices {
			s = strings.TrimSpace(s)
			if s != "" {
				allowed[s] = true
			}
		}
	}

	exempt := map[string]bool{
		// Whitelist gRPC reflection endpoints so Postman and grpcurl can inspect schema
		"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo":      true,
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": true,
	}

	for _, method := range exemptMethods {
		exempt[method] = true
	}

	return &AuthInterceptor{
		expectedToken:   expectedToken,
		allowedServices: allowed,
		exemptMethods:   exempt,
	}
}

// Unary returns a gRPC unary server interceptor for token & service authentication
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentication for exempt endpoints
		if a.exemptMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		newCtx, err := a.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// Stream returns a gRPC stream server interceptor for token & service authentication
func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if a.exemptMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		newCtx, err := a.authenticate(ss.Context())
		if err != nil {
			return err
		}

		_ = newCtx
		return handler(srv, ss)
	}
}

func (a *AuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata headers")
	}

	// 1. Verify Calling Service Name Header (e.g. X-Service-Name)
	serviceHeaders := md.Get("x-service-name")
	if len(serviceHeaders) == 0 || strings.TrimSpace(serviceHeaders[0]) == "" {
		return nil, status.Error(codes.PermissionDenied, "missing caller service identification (x-service-name)")
	}

	serviceName := strings.TrimSpace(serviceHeaders[0])
	if !a.allowedServices[serviceName] {
		return nil, status.Errorf(codes.PermissionDenied, "service '%s' is not authorized to access accounts service", serviceName)
	}

	// 2. Verify Bearer Token or Service Token
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		tokenHeaders := md.Get("x-service-token")
		if len(tokenHeaders) > 0 {
			authHeaders = tokenHeaders
		} else {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}
	}

	token := authHeaders[0]
	// Strip Bearer prefix if provided
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}

	if token != a.expectedToken {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication token")
	}

	ctx = context.WithValue(ctx, AuthTokenContextKey, token)
	ctx = context.WithValue(ctx, ServiceNameContextKey, serviceName)
	return ctx, nil
}
