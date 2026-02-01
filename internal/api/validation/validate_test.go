package validation

import (
	"testing"

	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

func TestValidatePosition(t *testing.T) {
	tests := []struct {
		name      string
		position  models.Position
		wantError bool
		errorCode string
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
			errorCode: "INVALID_LATITUDE",
		},
		{
			name: "Latitude too low",
			position: models.Position{
				Latitude:  -95.0,
				Longitude: 34.7818,
				Altitude:  1000.0,
			},
			wantError: true,
			errorCode: "INVALID_LATITUDE",
		},
		{
			name: "Longitude too high",
			position: models.Position{
				Latitude:  32.0853,
				Longitude: 185.0,
				Altitude:  1000.0,
			},
			wantError: true,
			errorCode: "INVALID_LONGITUDE",
		},
		{
			name: "Longitude too low",
			position: models.Position{
				Latitude:  32.0853,
				Longitude: -185.0,
				Altitude:  1000.0,
			},
			wantError: true,
			errorCode: "INVALID_LONGITUDE",
		},
		{
			name: "Negative altitude",
			position: models.Position{
				Latitude:  32.0853,
				Longitude: 34.7818,
				Altitude:  -100.0,
			},
			wantError: true,
			errorCode: "INVALID_ALTITUDE",
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
				
				if validationErr, ok := err.(*models.ValidationError); ok {
					if validationErr.Code != tt.errorCode {
						t.Errorf("ValidatePosition() error code = %v, want %v",
							validationErr.Code, tt.errorCode)
					}
				} else {
					t.Errorf("ValidatePosition() error type is not ValidationError")
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
		errorCode string
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
			errorCode: "INVALID_SPEED",
		},
		{
			name:      "Exceeds max speed",
			speed:     300.0,
			maxSpeed:  250.0,
			wantError: true,
			errorCode: "SPEED_EXCEEDS_MAX",
		},
		{
			name:      "Slightly exceeds max",
			speed:     250.1,
			maxSpeed:  250.0,
			wantError: true,
			errorCode: "SPEED_EXCEEDS_MAX",
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
				
				if validationErr, ok := err.(*models.ValidationError); ok {
					if validationErr.Code != tt.errorCode {
						t.Errorf("ValidateSpeed() error code = %v, want %v",
							validationErr.Code, tt.errorCode)
					}
				} else {
					t.Errorf("ValidateSpeed() error type is not ValidationError")
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
		errorCode string
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
			errorCode: "EMPTY_WAYPOINTS",
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
				
				if tt.errorCode != "" {
					if validationErr, ok := err.(*models.ValidationError); ok {
						if validationErr.Code != tt.errorCode {
							t.Errorf("ValidateTrajectoryCommand() error code = %v, want %v",
								validationErr.Code, tt.errorCode)
						}
					}
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
