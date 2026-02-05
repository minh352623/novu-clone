package response

import (
	"fmt"
	"net/http"
)

type APIError struct {
	StatusCode int
	Message    string
	Err        interface{}
}

func (e *APIError) Error() string {
	switch v := e.Err.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func NewAPIError(status int, message string, err interface{}) *APIError {
	return &APIError{
		StatusCode: status,
		Message:    message,
		Err:        err,
	}
}

// NewBadRequestError creates a 400 Bad Request error
func NewBadRequestError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
		Err:        message,
	}
}

// NewUnauthorizedError creates a 401 Unauthorized error
func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusUnauthorized,
		Message:    message,
		Err:        message,
	}
}

// NewForbiddenError creates a 403 Forbidden error
func NewForbiddenError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusForbidden,
		Message:    message,
		Err:        message,
	}
}

// NewNotFoundError creates a 404 Not Found error
func NewNotFoundError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusNotFound,
		Message:    message,
		Err:        message,
	}
}

// NewInternalServerError creates a 500 Internal Server Error
func NewInternalServerError(message string) *APIError {
	return &APIError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
		Err:        message,
	}
}
