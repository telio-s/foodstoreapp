package port

import "food-store-apis/internal/domain/apperror"

// AppError, Code and the constructors below re-export domain/apperror so
// adapters (e.g. adapter/http) construct and inspect application errors
// through port, the same way they consume domain/model, instead of
// importing the domain package directly.
type (
	AppError = apperror.AppError
	Code      = apperror.Code
)

const (
	CodeNotFound = apperror.CodeNotFound
	CodeInvalid  = apperror.CodeInvalid
	CodeInternal = apperror.CodeInternal
)

var (
	NotFound = apperror.NotFound
	Invalid  = apperror.Invalid
	Internal = apperror.Internal

	// ErrValidation is returned when an inbound HTTP request body fails DTO
	// validation. See domain/apperror for its ErrorCode (-9000) and usage.
	ErrValidation = apperror.ErrValidation
)
