package apperror

import "fmt"

type Code string

const (
	CodeNotFound Code = "NOT_FOUND"
	CodeInvalid  Code = "INVALID_INPUT"
	CodeInternal Code = "INTERNAL"
)

type AppError struct {
	Code      Code
	ErrorCode int
	Message   string
	Errors    []string
	Err       error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// WithErrors returns a copy of e carrying the given detail messages (e.g.
// per-field validation failures). It clones rather than mutates e so
// package-level sentinel errors such as ErrValidation stay safe to reuse
// across requests.
func (e *AppError) WithErrors(errs []string) *AppError {
	clone := *e
	clone.Errors = errs
	return &clone
}

func NotFound(message string) *AppError {
	return &AppError{Code: CodeNotFound, Message: message}
}

func Invalid(message string) *AppError {
	return &AppError{Code: CodeInvalid, Message: message}
}

// internalErrorCode is the single ErrorCode reported for every unexpected
// (CodeInternal) failure -- callers don't get to pick their own, since the
// client can't act on an internal error any differently than another.
const internalErrorCode = -9999

func Internal(err error) *AppError {
	return &AppError{Code: CodeInternal, ErrorCode: internalErrorCode, Message: "internal error", Err: err}
}

// Sentinel errors carrying a stable numeric ErrorCode for API consumers,
// independent of the string Code used for HTTP status mapping. Attach
// request-specific details with AppError.WithErrors, e.g.:
//
//	apperror.ErrValidation.WithErrors([]string{"product_id is required"})
//
// Expected (non-internal) business errors are numbered sequentially from
// -1000 as they're added; -9000 and -9999 are reserved for validation and
// internal failures respectively.
var (
	// ErrValidation is returned when an inbound HTTP request body fails DTO
	// validation, e.g. POST /api/v1/order missing a required field:
	//
	//	{ "code": -9000, "msg": "request validation failed", "error": ["product_id is required"] }
	ErrValidation = &AppError{Code: CodeInvalid, ErrorCode: -9000, Message: "request validation failed"}

	// ErrOrderEmptyItems is returned by OrderService.CreateOrder when the
	// order has no line items.
	ErrOrderEmptyItems = &AppError{Code: CodeInvalid, ErrorCode: -1000, Message: "order must contain at least one item"}

	// ErrProductNotFound is returned by OrderService.CreateOrder when an
	// order line references a product that doesn't exist.
	ErrProductNotFound = &AppError{Code: CodeNotFound, ErrorCode: -1001, Message: "product not found"}
)
