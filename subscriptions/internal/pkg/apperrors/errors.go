package apperrors

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrAlreadyExists     = errors.New("record already exists")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrInvalidStatus     = errors.New("invalid status transition")
	ErrInvoiceAlreadyPaid = errors.New("invoice is already paid")
	ErrInvoiceVoided     = errors.New("invoice is voided")
)
