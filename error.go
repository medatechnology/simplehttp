package simplehttp

import "fmt"

// Error types
var (
	ErrInvalidConfig     = fmt.Errorf("invalid configuration")
	ErrServerStartup     = fmt.Errorf("server startup failed")
	ErrNotFound          = fmt.Errorf("resource not found")
	ErrUnauthorized      = fmt.Errorf("unauthorized")
	ErrForbidden         = fmt.Errorf("forbidden")
	ErrRateLimitExceeded = fmt.Errorf("limit exceeded")
)

// SimpleHttpError represents a standardized error response
type SimpleHttpError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *SimpleHttpError) Error() string {
	return e.Message
}

// NewError creates a new SimpleHttpError
func NewError(code int, message string, details ...interface{}) *SimpleHttpError {
	var detailsData interface{}
	if len(details) > 0 {
		detailsData = details[0]
	}
	return &SimpleHttpError{
		Code:    code,
		Message: message,
		Details: detailsData,
	}
}
