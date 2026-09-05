package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	expectedToken   string
	allowedServices map[string]bool
}

func NewAuthInterceptor(expectedToken string, allowedServices []string) *AuthInterceptor {
	servicesMap := make(map[string]bool)
	for _, s := range allowedServices {
		servicesMap[strings.TrimSpace(s)] = true
	}

	return &AuthInterceptor{
		expectedToken:   expectedToken,
		allowedServices: servicesMap,
	}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := a.authorize(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := a.authorize(ss.Context()); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func (a *AuthInterceptor) authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "missing metadata context")
	}

	// 1. Verify Authorization Bearer Token
	if a.expectedToken != "" {
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return status.Errorf(codes.Unauthenticated, "authorization token required in header")
		}

		token := strings.TrimPrefix(authHeaders[0], "Bearer ")
		token = strings.TrimSpace(token)

		if token != a.expectedToken {
			return status.Errorf(codes.PermissionDenied, "invalid service authorization token")
		}
	}

	// 2. Verify Whitelisted Calling Service
	if len(a.allowedServices) > 0 {
		serviceHeaders := md.Get("x-service-name")
		if len(serviceHeaders) == 0 {
			return status.Errorf(codes.PermissionDenied, "caller service identity (x-service-name) required")
		}

		callerService := strings.TrimSpace(serviceHeaders[0])
		if !a.allowedServices[callerService] {
			return status.Errorf(codes.PermissionDenied, "service '%s' is not authorized to call webhook runner", callerService)
		}
	}

	return nil
}
