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

// ServiceAuthInterceptor validates caller service name and auth token
func ServiceAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		// 1. Verify Caller Service Name
		serviceNames := md.Get("x-service-name")
		if len(serviceNames) == 0 || strings.TrimSpace(serviceNames[0]) == "" {
			return nil, status.Errorf(codes.PermissionDenied, "missing x-service-name header")
		}

		callerService := strings.TrimSpace(serviceNames[0])
		allowedServicesEnv := os.Getenv("ALLOWED_SERVICES")
		if allowedServicesEnv == "" {
			allowedServicesEnv = "api-gateway,webhook-runner"
		}

		allowedList := strings.Split(allowedServicesEnv, ",")
		isAllowed := false
		for _, s := range allowedList {
			if strings.TrimSpace(s) == callerService {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return nil, status.Errorf(codes.PermissionDenied, "service '%s' is not authorized to access subscriptions", callerService)
		}

		// 2. Verify Auth Token
		expectedToken := os.Getenv("AUTH_TOKEN")
		if expectedToken != "" {
			authHeaders := md.Get("authorization")
			if len(authHeaders) == 0 {
				return nil, status.Errorf(codes.Unauthenticated, "missing authorization token")
			}

			token := strings.TrimPrefix(authHeaders[0], "Bearer ")
			token = strings.TrimSpace(token)

			if token != expectedToken {
				return nil, status.Errorf(codes.Unauthenticated, "invalid service authorization token")
			}
		}

		return handler(ctx, req)
	}
}
