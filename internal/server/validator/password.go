package validator

import (
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
)

// Register registers custom validators used by Gin binding.
func Register() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	_ = v.RegisterValidation("password", validatePassword)
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return IsValidPassword(password)
}

// IsValidPassword checks password complexity rules.
func IsValidPassword(password string) bool {
	if len(password) < MinPasswordLength || len(password) > MaxPasswordLength {
		return false
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}
