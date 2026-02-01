package models

import "time"

// AircraftState represents the complete state of the aircraft at a point in time.
type AircraftState struct {
	Position      Position          `json:"position"`
	Velocity      Velocity          `json:"velocity"`
	Heading       float64           `json:"heading"` // degrees, 0-360 (0=North)
	Timestamp     time.Time         `json:"timestamp"`
	ActiveCommand *CommandInfo      `json:"active_command,omitempty"`
	Environment   *EnvironmentState `json:"environment,omitempty"`
}

// Position represents geographic coordinates.
type Position struct {
	Latitude  float64 `json:"latitude"`  // degrees, -90 to 90
	Longitude float64 `json:"longitude"` // degrees, -180 to 180
	Altitude  float64 `json:"altitude"`  // meters MSL
}

// Velocity represents the aircraft's velocity vector.
type Velocity struct {
	GroundSpeed   float64 `json:"ground_speed"`   // m/s
	VerticalSpeed float64 `json:"vertical_speed"` // m/s (positive = climbing)
}

// CommandInfo contains information about the currently executing command.
type CommandInfo struct {
	Type       string    `json:"type"` // "goto", "trajectory", "hold", "stop"
	Target     *Position `json:"target,omitempty"`
	ETASeconds float64   `json:"eta_seconds,omitempty"`
}

// EnvironmentState represents environmental conditions.
type EnvironmentState struct {
	Wind     *WindVector `json:"wind,omitempty"`
	Humidity *float64    `json:"humidity,omitempty"` // 0-100%
}

// WindVector represents wind direction and speed.
type WindVector struct {
	Direction float64 `json:"direction"` // degrees
	Speed     float64 `json:"speed"`     // m/s
}
