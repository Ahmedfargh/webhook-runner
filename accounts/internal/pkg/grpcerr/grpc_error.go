package grpcerr

import (
	"errors"

	"accounts/internal/pkg/apperrors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCError translates domain and application errors to canonical gRPC status errors
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, apperrors.ErrNotFound), errors.Is(err, apperrors.ErrCountryNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, apperrors.ErrInvalidArgument), errors.Is(err, apperrors.ErrPhoneInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, apperrors.ErrAlreadyExists), errors.Is(err, apperrors.ErrEmailAlreadyUsed):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
