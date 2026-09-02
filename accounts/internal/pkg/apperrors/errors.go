package apperrors

import "errors"

var (
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrInternal         = errors.New("internal server error")
	ErrEmailAlreadyUsed = errors.New("email is already registered")
	ErrPhoneInvalid     = errors.New("phone number is invalid")
	ErrCountryNotFound  = errors.New("country not found")
)
