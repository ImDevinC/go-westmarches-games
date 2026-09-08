package westmarches

import "fmt"

// APIError represents an error response returned by the West Marches API.
// It wraps the HTTP status code and, when the server provided one, the
// application-level error message from the response body.
type APIError struct {
	// StatusCode is the HTTP status code returned by the server.
	StatusCode int
	// Success mirrors the "success" field of the error body.
	Success bool
	// Message mirrors the "error" field of the error body.
	Message string
	// RateLimit is the rate-limit status reported alongside the error, if any.
	RateLimit *RateLimit
}

// Error implements the error interface.
func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = httpStatusText(e.StatusCode)
	}
	return fmt.Sprintf("westmarches: API request failed (status %d): %s", e.StatusCode, msg)
}

// IsNotFound reports whether the error is a 404 Not Found response.
func (e *APIError) IsNotFound() bool { return e.StatusCode == httpStatusNotFound }

// IsUnauthorized reports whether the error is a 401 Unauthorized response.
func (e *APIError) IsUnauthorized() bool { return e.StatusCode == httpStatusUnauthorized }

// IsForbidden reports whether the error is a 403 Forbidden response.
func (e *APIError) IsForbidden() bool { return e.StatusCode == httpStatusForbidden }

// IsRateLimited reports whether the error is a 429 Too Many Requests response.
func (e *APIError) IsRateLimited() bool { return e.StatusCode == httpStatusTooManyRequests }

func httpStatusText(code int) string {
	switch code {
	case 400:
		return "bad request"
	case 401:
		return "unauthorized"
	case 403:
		return "forbidden"
	case 404:
		return "not found"
	case 429:
		return "too many requests"
	case 502:
		return "bad gateway"
	default:
		return "unknown error"
	}
}

const (
	httpStatusUnauthorized    = 401
	httpStatusForbidden       = 403
	httpStatusNotFound        = 404
	httpStatusTooManyRequests = 429
)
