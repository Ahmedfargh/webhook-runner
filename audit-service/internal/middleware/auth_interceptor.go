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

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := i.authorize(ctx, info.FullMethod); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := i.authorize(ss.Context(), info.FullMethod); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context, method string) error {
	if i.expectedToken == "" {
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	token := values[0]
	token = strings.TrimPrefix(token, "Bearer ")

	if token != i.expectedToken {
		return status.Errorf(codes.Unauthenticated, "invalid service authorization token")
	}

	if len(i.allowedServices) > 0 {
		serviceNames := md.Get("x-service-name")
		if len(serviceNames) == 0 {
			return status.Errorf(codes.PermissionDenied, "x-service-name header is required")
		}
		if !i.allowedServices[serviceNames[0]] {
			return status.Errorf(codes.PermissionDenied, "service %s is not authorized to access audit service", serviceNames[0])
		}
	}

	return nil
}
