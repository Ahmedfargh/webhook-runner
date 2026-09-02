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

func TestAuthInterceptor_TokenAuth(t *testing.T) {
	secretToken := "my-secret-token"
	interceptor := middleware.NewAuthInterceptor(secretToken)

	dummyHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		token := ctx.Value(middleware.AuthTokenContextKey)
		assert.Equal(t, secretToken, token)
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

	// 2. Invalid Token -> Unauthenticated
	ctxWithBadToken := metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer wrong-token"))
	_, err = interceptor.Unary()(ctxWithBadToken, nil, info, dummyHandler)
	assert.Error(t, err)
	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())

	// 3. Valid Token with Bearer prefix -> Success
	ctxWithGoodToken := metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer my-secret-token"))
	res, err := interceptor.Unary()(ctxWithGoodToken, nil, info, dummyHandler)
	require.NoError(t, err)
	assert.Equal(t, "success", res)

	// 4. Whitelisted Reflection method -> Success without header
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
