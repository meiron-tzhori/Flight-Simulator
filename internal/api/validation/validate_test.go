package validation

import (
	"errors"
	"strings"
	"testing"

	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

func TestValidatePosition(t *testing.T) {
	tests := []struct {
		name      string
		position  models.Position
		wantError bool
		errorType error
	}{
		{
			name: "Valid position",
			position: models.Position{
				Latitude:  32.0853,
				Longitude: 34.7818,
				Altitude:  1000.0,
			},
			wantError: false,
		},
		{
			name: "Valid at boundaries",
			position: models.Position{
				Latitude:  90.0,
				Longitude: 180.0,
				Altitude:  0.0,
			},
			wantError: false,
		},
		{
			name: "Valid negative boundaries",
			position: models.Position{
				Latitude:  -90.0,
				Longitude: -180.0,
				Altitude:  0.0,
			},
			wantError: false,
		},
		{
			name: "Latitude too high",
			position: models.Position{
				Latitude:  95.0,
				Longitude: 34.7818,
				Altitude:  1000.0,
			},
			wantError: true,
			errorType: models.ErrInvalidLatitude,
		},
		{
			name: "Latitude too low",
			position: models.Position{
				Latitude:  -95.0,
				Longitude: 34.7818,
				Altitude:  1000.0,
			},
			wantError: true,
			errorType: models.ErrInvalidLatitude,
		},
		{
			name: "Longitude too high",
			position: models.Position{
				Latitude:  32.0853,
				Longitude: 185.0,
				Altitude:  1000.0,
			},
			wantError: true,
			errorType: models.ErrInvalidLongitude,
		},
		{
			name: "Longitude too low",
			position: models.Position{
				Latitude:  32.0853,
				Longitude: -185.0,
				Altitude:  1000.0,
			},
			wantError: true,
			errorType: models.ErrInvalidLongitude,
		},
		{
			name: "Negative altitude",
			position: models.Position{
				Latitude:  32.0853,
				Longitude: 34.7818,
				Altitude:  -100.0,
			},
			wantError: true,
			errorType: models.ErrInvalidAltitude,
		},
		{
			name: "Very high altitude (valid)",
			position: models.Position{
				Latitude:  32.0853,
				Longitude: 34.7818,
				Altitude:  50000.0,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePosition(tt.position)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidatePosition() expected error, got nil")
					return
				}

				// Check error contains expected type
				if !errors.Is(err, tt.errorType) {
					t.Errorf("ValidatePosition() error = %v, want to contain %v", err, tt.errorType)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePosition() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateSpeed(t *testing.T) {
	tests := []struct {
		name      string
		speed     float64
		maxSpeed  float64
		wantError bool
		errorType error
	}{
		{
			name:      "Valid speed",
			speed:     50.0,
			maxSpeed:  250.0,
			wantError: false,
		},
		{
			name:      "Zero speed (valid)",
			speed:     0.0,
			maxSpeed:  250.0,
			wantError: false,
		},
		{
			name:      "Max speed",
			speed:     250.0,
			maxSpeed:  250.0,
			wantError: false,
		},
		{
			name:      "Negative speed",
			speed:     -10.0,
			maxSpeed:  250.0,
			wantError: true,
			errorType: models.ErrInvalidSpeed,
		},
		{
			name:      "Exceeds max speed",
			speed:     300.0,
			maxSpeed:  250.0,
			wantError: true,
			errorType: models.ErrSpeedExceedsMax,
		},
		{
			name:      "Slightly exceeds max",
			speed:     250.1,
			maxSpeed:  250.0,
			wantError: true,
			errorType: models.ErrSpeedExceedsMax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpeed(tt.speed, tt.maxSpeed)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateSpeed() expected error, got nil")
					return
				}

				// Check error contains expected type
				if !errors.Is(err, tt.errorType) {
					t.Errorf("ValidateSpeed() error = %v, want to contain %v", err, tt.errorType)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateSpeed() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateGoToCommand(t *testing.T) {
	tests := []struct {
		name      string
		cmd       *models.GoToCommand
		maxSpeed  float64
		wantError bool
	}{
		{
			name: "Valid command without speed",
			cmd: &models.GoToCommand{
				Target: models.Position{
					Latitude:  32.0853,
					Longitude: 34.7818,
					Altitude:  1000.0,
				},
			},
			maxSpeed:  250.0,
			wantError: false,
		},
		{
			name: "Valid command with speed",
			cmd: &models.GoToCommand{
				Target: models.Position{
					Latitude:  32.0853,
					Longitude: 34.7818,
					Altitude:  1000.0,
				},
				Speed: ptr(50.0),
			},
			maxSpeed:  250.0,
			wantError: false,
		},
		{
			name: "Invalid target position",
			cmd: &models.GoToCommand{
				Target: models.Position{
					Latitude:  95.0,
					Longitude: 34.7818,
					Altitude:  1000.0,
				},
			},
			maxSpeed:  250.0,
			wantError: true,
		},
		{
			name: "Invalid speed",
			cmd: &models.GoToCommand{
				Target: models.Position{
					Latitude:  32.0853,
					Longitude: 34.7818,
					Altitude:  1000.0,
				},
				Speed: ptr(300.0),
			},
			maxSpeed:  250.0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGoToCommand(tt.cmd, tt.maxSpeed)

			if tt.wantError && err == nil {
				t.Errorf("ValidateGoToCommand() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateGoToCommand() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateTrajectoryCommand(t *testing.T) {
	tests := []struct {
		name      string
		cmd       *models.TrajectoryCommand
		maxSpeed  float64
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid trajectory",
			cmd: &models.TrajectoryCommand{
				Waypoints: []models.Waypoint{
					{
						Position: models.Position{Latitude: 32.0, Longitude: 34.7, Altitude: 1000},
						Speed:    ptr(50.0),
					},
					{
						Position: models.Position{Latitude: 32.1, Longitude: 34.8, Altitude: 1500},
						Speed:    ptr(100.0),
					},
				},
				Loop: false,
			},
			maxSpeed:  250.0,
			wantError: false,
		},
		{
			name: "Empty waypoints",
			cmd: &models.TrajectoryCommand{
				Waypoints: []models.Waypoint{},
				Loop:      false,
			},
			maxSpeed:  250.0,
			wantError: true,
			errorMsg:  "waypoint",
		},
		{
			name: "Invalid waypoint position",
			cmd: &models.TrajectoryCommand{
				Waypoints: []models.Waypoint{
					{
						Position: models.Position{Latitude: 95.0, Longitude: 34.7, Altitude: 1000},
					},
				},
				Loop: false,
			},
			maxSpeed:  250.0,
			wantError: true,
			errorMsg:  "latitude",
		},
		{
			name: "Invalid waypoint speed",
			cmd: &models.TrajectoryCommand{
				Waypoints: []models.Waypoint{
					{
						Position: models.Position{Latitude: 32.0, Longitude: 34.7, Altitude: 1000},
						Speed:    ptr(300.0),
					},
				},
				Loop: false,
			},
			maxSpeed:  250.0,
			wantError: true,
			errorMsg:  "speed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTrajectoryCommand(tt.cmd, tt.maxSpeed)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateTrajectoryCommand() expected error, got nil")
					return
				}

				// Check error message contains expected text
				if tt.errorMsg != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errorMsg)) {
					t.Errorf("ValidateTrajectoryCommand() error = %v, want to contain %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTrajectoryCommand() unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to create pointer to float64
func ptr(f float64) *float64 {
	return &f
}
