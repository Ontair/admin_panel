package dto

import (
	"net/http"
)

// APIError represents API error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements error interface
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError creates new API error
func NewAPIError(code int, message string, details string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common API errors
var (
	// HTTP 400
	ErrBadRequest         = NewAPIError(http.StatusBadRequest, "Bad Request", "")
	ErrValidationFailed   = NewAPIError(http.StatusBadRequest, "Validation Failed", "")
	ErrInvalidCredentials = NewAPIError(http.StatusBadRequest, "Invalid credentials", "")

	// HTTP 401
	ErrUnauthorized = NewAPIError(http.StatusUnauthorized, "Unauthorized", "")
	ErrInvalidToken = NewAPIError(http.StatusUnauthorized, "Invalid token", "")
	ErrTokenExpired = NewAPIError(http.StatusUnauthorized, "Token expired", "")

	// HTTP 403
	ErrForbidden              = NewAPIError(http.StatusForbidden, "Forbidden", "")
	ErrInsufficientPrivileges = NewAPIError(http.StatusForbidden, "Insufficient privileges", "")

	// HTTP 404
	ErrNotFound     = NewAPIError(http.StatusNotFound, "Not Found", "")
	ErrUserNotFound = NewAPIError(http.StatusNotFound, "User not found", "")

	// HTTP 409
	ErrConflict          = NewAPIError(http.StatusConflict, "Conflict", "")
	ErrUserAlreadyExists = NewAPIError(http.StatusConflict, "User already exists", "")

	// HTTP 422
	ErrUnprocessableEntity = NewAPIError(http.StatusUnprocessableEntity, "Unprocessable Entity", "")

	// HTTP 500
	ErrInternalServer = NewAPIError(http.StatusInternalServerError, "Internal Server Error", "")
	ErrDatabaseError  = NewAPIError(http.StatusInternalServerError, "Database Error", "")
)
