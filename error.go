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

// MedaError represents a standardized error response
type MedaError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *MedaError) Error() string {
	return e.Message
}

// NewError creates a new MedaError
func NewError(code int, message string, details ...interface{}) *MedaError {
	var detailsData interface{}
	if len(details) > 0 {
		detailsData = details[0]
	}
	return &MedaError{
		Code:    code,
		Message: message,
		Details: detailsData,
	}
}
