package validation

import (
	"fmt"

	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

// ValidatePosition validates geographic coordinates.
func ValidatePosition(pos models.Position) error {
	if pos.Latitude < -90 || pos.Latitude > 90 {
		return fmt.Errorf("%w: %f", models.ErrInvalidLatitude, pos.Latitude)
	}
	if pos.Longitude < -180 || pos.Longitude > 180 {
		return fmt.Errorf("%w: %f", models.ErrInvalidLongitude, pos.Longitude)
	}
	if pos.Altitude < 0 {
		return fmt.Errorf("%w: %f", models.ErrInvalidAltitude, pos.Altitude)
	}
	return nil
}

// ValidateSpeed validates speed value.
func ValidateSpeed(speed float64, maxSpeed float64) error {
	if speed < 0 {
		return models.ErrInvalidSpeed
	}
	if speed > maxSpeed {
		return fmt.Errorf("%w: %f > %f", models.ErrSpeedExceedsMax, speed, maxSpeed)
	}
	return nil
}

// ValidateGoToCommand validates a go-to command.
func ValidateGoToCommand(cmd *models.GoToCommand, maxSpeed float64) error {
	if err := ValidatePosition(cmd.Target); err != nil {
		return err
	}
	if cmd.Speed != nil {
		if err := ValidateSpeed(*cmd.Speed, maxSpeed); err != nil {
			return err
		}
	}
	return nil
}

// ValidateTrajectoryCommand validates a trajectory command.
func ValidateTrajectoryCommand(cmd *models.TrajectoryCommand, maxSpeed float64) error {
	if len(cmd.Waypoints) == 0 {
		return models.ErrEmptyWaypoints
	}

	for i, wp := range cmd.Waypoints {
		if err := ValidatePosition(wp.Position); err != nil {
			return fmt.Errorf("waypoint %d: %w", i, err)
		}
		if wp.Speed != nil {
			if err := ValidateSpeed(*wp.Speed, maxSpeed); err != nil {
				return fmt.Errorf("waypoint %d: %w", i, err)
			}
		}
	}

	return nil
}
