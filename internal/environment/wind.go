package environment

import (
	"math"

	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

// WindEffect applies wind to aircraft velocity.
type WindEffect struct {
	direction float64 // degrees (0-360, where 0 is North)
	speed     float64 // m/s
}

// NewWindEffect creates a new wind effect.
func NewWindEffect(direction, speed float64) *WindEffect {
	return &WindEffect{
		direction: direction,
		speed:     speed,
	}
}

// Apply applies wind effect to velocity, returning the effective ground velocity.
// The aircraft maintains its airspeed and heading, but wind affects ground speed and track.
func (w *WindEffect) Apply(heading float64, velocity models.Velocity) models.Velocity {
	// Convert to radians
	headingRad := heading * math.Pi / 180.0
	windDirRad := w.direction * math.Pi / 180.0

	// Calculate aircraft velocity components (airspeed)
	// In aviation, heading is the direction the aircraft is pointing
	acNorth := velocity.GroundSpeed * math.Cos(headingRad)
	acEast := velocity.GroundSpeed * math.Sin(headingRad)

	// Calculate wind velocity components
	// Wind direction is "from" direction, so we need to add 180° or use opposite signs
	// If wind is "from North" (0°), it pushes South (negative North component)
	windNorth := -w.speed * math.Cos(windDirRad)
	windEast := -w.speed * math.Sin(windDirRad)

	// Add wind effect to get ground velocity
	groundNorth := acNorth + windNorth
	groundEast := acEast + windEast

	// Calculate new ground speed (magnitude of ground velocity vector)
	newGroundSpeed := math.Sqrt(groundNorth*groundNorth + groundEast*groundEast)

	// Vertical speed is not affected by horizontal wind
	return models.Velocity{
		GroundSpeed:   newGroundSpeed,
		VerticalSpeed: velocity.VerticalSpeed,
	}
}

// GetVector returns the wind vector for reporting.
func (w *WindEffect) GetVector() *models.WindVector {
	return &models.WindVector{
		Direction: w.direction,
		Speed:     w.speed,
	}
}

// CalculateHeadwindComponent calculates the headwind component for a given heading.
// Positive values indicate headwind, negative values indicate tailwind.
func (w *WindEffect) CalculateHeadwindComponent(heading float64) float64 {
	// Convert to radians
	headingRad := heading * math.Pi / 180.0
	windDirRad := w.direction * math.Pi / 180.0

	// Calculate wind components
	windNorth := -w.speed * math.Cos(windDirRad)
	windEast := -w.speed * math.Sin(windDirRad)

	// Calculate aircraft heading components
	acNorth := math.Cos(headingRad)
	acEast := math.Sin(headingRad)

	// Dot product gives headwind component
	// Positive = headwind, Negative = tailwind
	headwind := -(windNorth*acNorth + windEast*acEast)

	return headwind
}

// CalculateCrosswindComponent calculates the crosswind component for a given heading.
// Positive values indicate wind from the right, negative from the left.
func (w *WindEffect) CalculateCrosswindComponent(heading float64) float64 {
	// Convert to radians
	headingRad := heading * math.Pi / 180.0
	windDirRad := w.direction * math.Pi / 180.0

	// Calculate wind components
	windNorth := -w.speed * math.Cos(windDirRad)
	windEast := -w.speed * math.Sin(windDirRad)

	// Calculate aircraft perpendicular components (right side)
	acRight := math.Sin(headingRad)
	acLeft := -math.Cos(headingRad)

	// Cross product magnitude gives crosswind component
	crosswind := windNorth*acRight + windEast*acLeft

	return crosswind
}
