package validator

import (
	"errors"
	"fmt"
	"strings"

	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/go-playground/validator/v10"
)

// BindError converts Gin/validator binding errors into a user-facing AppError.
func BindError(err error) appErrors.AppError {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		messages := make([]string, 0, len(validationErrs))
		for _, fieldErr := range validationErrs {
			messages = append(messages, fieldMessage(fieldErr))
		}
		return appErrors.ErrInvalidArgument.Reform(strings.Join(messages, "; "))
	}

	return appErrors.ErrInvalidArgument.Reform("invalid request body")
}

func fieldMessage(fieldErr validator.FieldError) string {
	field := humanField(fieldErr.Field())

	switch fieldErr.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "password":
		return fmt.Sprintf("%s must be %d-%d characters and include uppercase, lowercase, and a number", field, MinPasswordLength, MaxPasswordLength)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fieldErr.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fieldErr.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func humanField(field string) string {
	switch field {
	case "Email":
		return "Email"
	case "Password":
		return "Password"
	case "OldPassword":
		return "Old password"
	case "NewPassword":
		return "New password"
	case "ConfirmPassword":
		return "Confirm password"
	case "FirstName":
		return "First name"
	case "LastName":
		return "Last name"
	case "FullName":
		return "Full name"
	case "Role":
		return "Role"
	case "Status":
		return "Status"
	default:
		return field
	}
}
