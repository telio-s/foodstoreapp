package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"food-store-apis/internal/port"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// bindJSON decodes the request body into req and validates it with v (built
// by main.NewValidator and injected into the handler that calls this),
// returning port.ErrValidation with one human-readable message per failed
// field when validation fails.
func bindJSON(c *gin.Context, v *validator.Validate, req interface{}) *port.AppError {
	if err := json.NewDecoder(c.Request.Body).Decode(req); err != nil {
		return port.Invalid(err.Error())
	}

	if err := v.Struct(req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return port.ErrValidation.WithErrors(formatValidationErrors(validationErrs))
		}
		return port.Invalid(err.Error())
	}

	return nil
}

func formatValidationErrors(validationErrs validator.ValidationErrors) []string {
	messages := make([]string, 0, len(validationErrs))
	for _, fe := range validationErrs {
		messages = append(messages, formatFieldError(fe))
	}
	return messages
}

func formatFieldError(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
