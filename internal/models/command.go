package models

import "github.com/google/uuid"

// CommandType represents the type of command.
type CommandType string

const (
	CommandTypeGoTo       CommandType = "goto"
	CommandTypeTrajectory CommandType = "trajectory"
	CommandTypeStop       CommandType = "stop"
	CommandTypeHold       CommandType = "hold"
)

// Command represents a command to the aircraft.
type Command struct {
	ID         string           `json:"id"`
	Type       CommandType      `json:"type"`
	GoTo       *GoToCommand     `json:"goto,omitempty"`
	Trajectory *TrajectoryCommand `json:"trajectory,omitempty"`
}

// GoToCommand directs the aircraft to a specific point.
type GoToCommand struct {
	Target Position `json:"target"`
	Speed  *float64 `json:"speed,omitempty"` // m/s, optional
}

// TrajectoryCommand directs the aircraft to follow a sequence of waypoints.
type TrajectoryCommand struct {
	Waypoints []Waypoint `json:"waypoints"`
	Loop      bool       `json:"loop"`
}

// Waypoint represents a point in a trajectory.
type Waypoint struct {
	Position Position `json:"position"`
	Speed    *float64 `json:"speed,omitempty"` // m/s, optional
}

// NewCommand creates a new command with a unique ID.
func NewCommand(cmdType CommandType) *Command {
	return &Command{
		ID:   uuid.New().String(),
		Type: cmdType,
	}
}
