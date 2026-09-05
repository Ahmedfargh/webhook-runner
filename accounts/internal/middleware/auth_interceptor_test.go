package middleware_test

import (
	"context"
	"testing"

	"accounts/internal/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptor_ServiceAndTokenAuth(t *testing.T) {
	secretToken := "my-secret-token"
	allowedServices := []string{"api-gateway", "webhook-runner"}
	interceptor := middleware.NewAuthInterceptor(secretToken, allowedServices)

	dummyHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		token := ctx.Value(middleware.AuthTokenContextKey)
		service := ctx.Value(middleware.ServiceNameContextKey)
		assert.Equal(t, secretToken, token)
		assert.Equal(t, "api-gateway", service)
		return "success", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/accounts.v1.UserService/CreateUser",
	}

	// 1. Missing metadata -> Unauthenticated
	ctx := context.Background()
	_, err := interceptor.Unary()(ctx, nil, info, dummyHandler)
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())

	// 2. Missing Service Name Header -> PermissionDenied
	ctxNoService := metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer my-secret-token"))
	_, err = interceptor.Unary()(ctxNoService, nil, info, dummyHandler)
	assert.Error(t, err)
	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())

	// 3. Disallowed Service Name -> PermissionDenied
	ctxBadService := metadata.NewIncomingContext(ctx, metadata.Pairs(
		"x-service-name", "unauthorized-external-app",
		"authorization", "Bearer my-secret-token",
	))
	_, err = interceptor.Unary()(ctxBadService, nil, info, dummyHandler)
	assert.Error(t, err)
	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())

	// 4. Invalid Token -> Unauthenticated
	ctxWithBadToken := metadata.NewIncomingContext(ctx, metadata.Pairs(
		"x-service-name", "api-gateway",
		"authorization", "Bearer wrong-token",
	))
	_, err = interceptor.Unary()(ctxWithBadToken, nil, info, dummyHandler)
	assert.Error(t, err)
	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())

	// 5. Valid Service + Valid Token with Bearer prefix -> Success
	ctxWithGoodCreds := metadata.NewIncomingContext(ctx, metadata.Pairs(
		"x-service-name", "api-gateway",
		"authorization", "Bearer my-secret-token",
	))
	res, err := interceptor.Unary()(ctxWithGoodCreds, nil, info, dummyHandler)
	require.NoError(t, err)
	assert.Equal(t, "success", res)

	// 6. Valid Service + Valid x-service-token header -> Success
	ctxWithServiceToken := metadata.NewIncomingContext(ctx, metadata.Pairs(
		"x-service-name", "api-gateway",
		"x-service-token", "my-secret-token",
	))
	res, err = interceptor.Unary()(ctxWithServiceToken, nil, info, dummyHandler)
	require.NoError(t, err)
	assert.Equal(t, "success", res)

	// 7. Whitelisted Reflection method -> Success without headers
	reflectionInfo := &grpc.UnaryServerInfo{
		FullMethod: "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
	}
	reflectionHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "reflection-ok", nil
	}
	res, err = interceptor.Unary()(ctx, nil, reflectionInfo, reflectionHandler)
	require.NoError(t, err)
	assert.Equal(t, "reflection-ok", res)
}
