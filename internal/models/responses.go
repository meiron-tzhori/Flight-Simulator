package models

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status            string  `json:"status"`
	SimulationRunning bool    `json:"simulation_running"`
	TickRateHz        float64 `json:"tick_rate_hz"`
	UptimeSeconds     float64 `json:"uptime_seconds"`
	Timestamp         string  `json:"timestamp"`
}

// CommandResponse represents the response to a command submission.
type CommandResponse struct {
	Status         string    `json:"status"`
	CommandID      string    `json:"command_id"`
	Message        string    `json:"message"`
	Target         *Position `json:"target,omitempty"`
	WaypointCount  int       `json:"waypoint_count,omitempty"`
	ETASeconds     float64   `json:"eta_seconds,omitempty"`
	HoldPosition   *Position `json:"hold_position,omitempty"`
	OrbitRadiusM   float64   `json:"orbit_radius_meters,omitempty"`
}
