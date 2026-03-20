package mapbox

import (
	"encoding/json"
	"fmt"
)

// APIError represents an error response from the Mapbox API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mapbox API error (status %d): %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error represents a 404 Not Found response.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// checkResponse inspects the HTTP status code and returns an error for non-2xx responses.
func checkResponse(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	apiErr := &APIError{
		StatusCode: statusCode,
	}

	var errResp struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
		apiErr.Message = errResp.Message
	} else {
		apiErr.Message = string(body)
	}

	return apiErr
}

// IsNotFoundError checks if an error is a Mapbox API 404 Not Found error.
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsNotFound()
	}
	return false
}
