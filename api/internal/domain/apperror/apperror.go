package apperror

import "fmt"

type Code string

const (
	CodeNotFound Code = "NOT_FOUND"
	CodeInvalid  Code = "INVALID_INPUT"
	CodeInternal Code = "INTERNAL"
)

type AppError struct {
	Code    Code
	Message string
	Err     error
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

func NotFound(message string) *AppError {
	return &AppError{Code: CodeNotFound, Message: message}
}

func Invalid(message string) *AppError {
	return &AppError{Code: CodeInvalid, Message: message}
}

func Internal(err error) *AppError {
	return &AppError{Code: CodeInternal, Message: "internal error", Err: err}
}
