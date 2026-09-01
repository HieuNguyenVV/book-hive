package errors

import "net/http"

var (
	// 400 Bad Request
	ErrBadRequest = NewAppError(400000,
		"ErrBadRequest", http.StatusBadRequest, "bad request", "bad request")
	ErrInvalidArgument = NewAppError(400001,
		"ErrInvalidArgument", http.StatusBadRequest, "invalid argument", "invalid argument")
	ErrAlreadyExists = NewAppError(400002,
		"ErrAlreadyExists", http.StatusBadRequest, "already exists", "already exists")
	ErrInvalidValue = NewAppError(400003,
		"ErrInvalidValue", http.StatusBadRequest, "invalid value", "invalid value")
	ErrInvalidRequest = NewAppError(400004,
		"ErrInvalidRequest", http.StatusBadRequest, "invalid request", "invalid request")

	// 401 Unauthorized
	EErrInvalidSession = NewAppError(401000, "ErrInvalidSession", http.StatusUnauthorized, "Your session has expired or is invalid.", "Your session has expired or is invalid.")
	ErrTokenExpired    = NewAppError(401001,
		"ErrTokenExpired", http.StatusUnauthorized, "token expired", "token expired")
	ErrTokenNotValidYet = NewAppError(401002,
		"ErrTokenNotValidYet", http.StatusUnauthorized, "token not valid yet", "token not valid yet")
	ErrInvalidToken = NewAppError(401003,
		"ErrInvalidToken", http.StatusUnauthorized, "invalid token", "invalid token")
	ErrNilClaims = NewAppError(401004,
		"ErrNilClaims", http.StatusUnauthorized, "claims is nil", "claims is nil")
	ErrUnauthorized = NewAppError(401005,
		"ErrUnauthorized", http.StatusUnauthorized, "unauthorized", "unauthorized")

	// 403	Forbidden
	ErrForbidden          = NewAppError(403000, "ErrForbidden", http.StatusForbidden, "forbidden", "forbidden")
	ErrIPIsNotAllowed     = NewAppError(403001, "ErrIPIsNotAllowed", http.StatusForbidden, "IP is not allowed", "IP is not allowed")
	ErrDomainIsNotAllowed = NewAppError(403002, "ErrDomainIsNotAllowed", http.StatusForbidden, "domain is not allowed", "domain is not allowed")

	// 404 Not Found
	ErrNotFound     = NewAppError(404000, "ErrNotFound", http.StatusNotFound, "not found", "not found")
	ErrUserNotFound = NewAppError(404001, "ErrUserNotFound", http.StatusNotFound, "user not found", "user not found")

	// 500 Internal Server Error
	ErrInternalServerError = NewAppError(500000, "ErrInternalServerError", http.StatusInternalServerError, "internal server error", "internal server error")
	ErrDatabase            = NewAppError(500002,
		"ErrDatabase", http.StatusInternalServerError, "database error", "database error")
)
