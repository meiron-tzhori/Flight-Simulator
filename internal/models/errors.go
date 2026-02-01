package models

import "errors"

// Validation errors
var (
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90 degrees")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180 degrees")
	ErrInvalidAltitude  = errors.New("altitude must be non-negative")
	ErrInvalidSpeed     = errors.New("speed must be positive")
	ErrEmptyWaypoints   = errors.New("trajectory must contain at least one waypoint")
	ErrInvalidWaypoint  = errors.New("invalid waypoint")
	ErrSpeedExceedsMax  = errors.New("speed exceeds maximum allowed")
)

// Runtime errors
var (
	ErrCommandQueueFull    = errors.New("command queue is full")
	ErrSimulatorNotRunning = errors.New("simulator is not running")
	ErrTimeout             = errors.New("operation timeout")
	ErrTerrainConflict     = errors.New("terrain collision detected")
)

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains detailed error information.
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Field   string                 `json:"field,omitempty"`
	Value   interface{}            `json:"value,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}
