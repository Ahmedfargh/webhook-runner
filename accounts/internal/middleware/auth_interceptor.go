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
)

// AuthInterceptor enforces token-based authentication on incoming gRPC requests
type AuthInterceptor struct {
	expectedToken  string
	exemptMethods  map[string]bool
}

// NewAuthInterceptor creates a new token-based gRPC auth interceptor
func NewAuthInterceptor(expectedToken string, exemptMethods ...string) *AuthInterceptor {
	if expectedToken == "" {
		expectedToken = os.Getenv("AUTH_TOKEN")
		if expectedToken == "" {
			expectedToken = "webhook-accounts-secret-token"
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
		expectedToken: expectedToken,
		exemptMethods: exempt,
	}
}

// Unary returns a gRPC unary server interceptor for token authentication
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

// Stream returns a gRPC stream server interceptor for token authentication
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

		_, err := a.authenticate(ss.Context())
		if err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

func (a *AuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata headers")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := authHeaders[0]
	// Strip Bearer prefix if provided
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}

	if token != a.expectedToken {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication token")
	}

	return context.WithValue(ctx, AuthTokenContextKey, token), nil
}
