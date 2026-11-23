// Copied from github.com/daytonaio/daytona/libs/sdk-go/pkg/errors v0.190.0 (Apache-2.0); see NOTICE.

package errors

import (
	json "encoding/json"
	fmt "fmt"
	http "net/http"
)

// DaytonaError is the base error type for all Daytona SDK errors
type DaytonaError struct {
	Message    string
	StatusCode int
	Headers    http.Header
}

func (e *DaytonaError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("Daytona error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Daytona error: %s", e.Message)
}

// NewDaytonaError creates a new DaytonaError
func NewDaytonaError(message string, statusCode int, headers http.Header) *DaytonaError {
	return &DaytonaError{
		Message:    message,
		StatusCode: statusCode,
		Headers:    headers,
	}
}

// DaytonaNotFoundError represents a resource not found error (404)
type DaytonaNotFoundError struct {
	*DaytonaError
}

func (e *DaytonaNotFoundError) Error() string {
	return fmt.Sprintf("Resource not found: %s", e.Message)
}

// NewDaytonaNotFoundError creates a new DaytonaNotFoundError
func NewDaytonaNotFoundError(message string, headers http.Header) *DaytonaNotFoundError {
	return &DaytonaNotFoundError{
		DaytonaError: NewDaytonaError(message, http.StatusNotFound, headers),
	}
}

// DaytonaRateLimitError represents a rate limit error (429)
type DaytonaRateLimitError struct {
	*DaytonaError
}

func (e *DaytonaRateLimitError) Error() string {
	return fmt.Sprintf("Rate limit exceeded: %s", e.Message)
}

// NewDaytonaRateLimitError creates a new DaytonaRateLimitError
func NewDaytonaRateLimitError(message string, headers http.Header) *DaytonaRateLimitError {
	return &DaytonaRateLimitError{
		DaytonaError: NewDaytonaError(message, http.StatusTooManyRequests, headers),
	}
}

// DaytonaAuthenticationError represents an authentication error (401)
type DaytonaAuthenticationError struct {
	*DaytonaError
}

func (e *DaytonaAuthenticationError) Error() string {
	return fmt.Sprintf("Authentication failed: %s", e.Message)
}

func NewDaytonaAuthenticationError(message string, headers http.Header) *DaytonaAuthenticationError {
	return &DaytonaAuthenticationError{
		DaytonaError: NewDaytonaError(message, http.StatusUnauthorized, headers),
	}
}

// DaytonaForbiddenError represents a forbidden/authorization error (403)
type DaytonaForbiddenError struct {
	*DaytonaError
}

func (e *DaytonaForbiddenError) Error() string {
	return fmt.Sprintf("Forbidden: %s", e.Message)
}

func NewDaytonaForbiddenError(message string, headers http.Header) *DaytonaForbiddenError {
	return &DaytonaForbiddenError{
		DaytonaError: NewDaytonaError(message, http.StatusForbidden, headers),
	}
}

// DaytonaConflictError represents a conflict error (409)
type DaytonaConflictError struct {
	*DaytonaError
}

func (e *DaytonaConflictError) Error() string {
	return fmt.Sprintf("Conflict: %s", e.Message)
}

func NewDaytonaConflictError(message string, headers http.Header) *DaytonaConflictError {
	return &DaytonaConflictError{
		DaytonaError: NewDaytonaError(message, http.StatusConflict, headers),
	}
}

// DaytonaValidationError represents a validation/bad request error (400)
type DaytonaValidationError struct {
	*DaytonaError
}

func (e *DaytonaValidationError) Error() string {
	return fmt.Sprintf("Validation error: %s", e.Message)
}

func NewDaytonaValidationError(message string, headers http.Header) *DaytonaValidationError {
	return &DaytonaValidationError{
		DaytonaError: NewDaytonaError(message, http.StatusBadRequest, headers),
	}
}

// DaytonaServerError represents a server error (5xx)
type DaytonaServerError struct {
	*DaytonaError
}

func (e *DaytonaServerError) Error() string {
	return fmt.Sprintf("Server error: %s", e.Message)
}

func NewDaytonaServerError(message string, statusCode int, headers http.Header) *DaytonaServerError {
	return &DaytonaServerError{
		DaytonaError: NewDaytonaError(message, statusCode, headers),
	}
}

// DaytonaTimeoutError represents a timeout error
type DaytonaTimeoutError struct {
	*DaytonaError
}

func (e *DaytonaTimeoutError) Error() string {
	return fmt.Sprintf("Operation timed out: %s", e.Message)
}

func NewDaytonaTimeoutError(message string) *DaytonaTimeoutError {
	return &DaytonaTimeoutError{
		DaytonaError: NewDaytonaError(message, 0, nil),
	}
}

// NewDaytonaErrorFromBody parses a JSON response body and maps the status code
// to the appropriate SDK error type. Falls back to the raw body as the message.
func NewDaytonaErrorFromBody(body []byte, statusCode int, headers http.Header) error {
	var message string

	if len(body) > 0 {
		var errResp struct {
			Message    string `json:"message"`
			Error      string `json:"error"`
			StatusCode int    `json:"statusCode"`
		}
		if json.Unmarshal(body, &errResp) == nil {
			if errResp.Message != "" {
				message = errResp.Message
			} else if errResp.Error != "" {
				message = errResp.Error
			}
			if errResp.StatusCode != 0 {
				statusCode = errResp.StatusCode
			}
		}
		if message == "" {
			message = string(body)
		}
	}

	if message == "" {
		message = "Download failed"
	}

	switch statusCode {
	case http.StatusNotFound:
		return NewDaytonaNotFoundError(message, headers)
	case http.StatusTooManyRequests:
		return NewDaytonaRateLimitError(message, headers)
	default:
		return NewDaytonaError(message, statusCode, headers)
	}
}

func mapStatusCodeToError(statusCode int, message string, headers http.Header) error {
	switch {
	case statusCode == http.StatusBadRequest:
		return NewDaytonaValidationError(message, headers)
	case statusCode == http.StatusUnauthorized:
		return NewDaytonaAuthenticationError(message, headers)
	case statusCode == http.StatusForbidden:
		return NewDaytonaForbiddenError(message, headers)
	case statusCode == http.StatusNotFound:
		return NewDaytonaNotFoundError(message, headers)
	case statusCode == http.StatusConflict:
		return NewDaytonaConflictError(message, headers)
	case statusCode == http.StatusTooManyRequests:
		return NewDaytonaRateLimitError(message, headers)
	case statusCode >= 500 && statusCode <= 599:
		return NewDaytonaServerError(message, statusCode, headers)
	case statusCode == 0:
		return NewDaytonaError(message, 0, nil)
	default:
		return NewDaytonaError(message, statusCode, headers)
	}
}
