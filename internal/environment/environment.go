package environment

import (
	"github.com/meiron-tzhori/Flight-Simulator/internal/config"
	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

// Environment manages environmental effects on the aircraft.
type Environment struct {
	wind     *WindEffect
	humidity *float64
	enabled  bool
}

// New creates a new environment from configuration.
func New(cfg config.EnvironmentConfig) *Environment {
	if !cfg.Enabled {
		return nil
	}

	env := &Environment{
		enabled: true,
	}

	// Initialize wind if enabled
	if cfg.Wind.Enabled {
		env.wind = NewWindEffect(cfg.Wind.Direction, cfg.Wind.Speed)
	}

	// Initialize humidity if enabled
	if cfg.Humidity.Enabled {
		env.humidity = &cfg.Humidity.Value
	}

	return env
}

// ApplyEffects applies all enabled environmental effects to the velocity.
// Returns the effective velocity after environmental effects.
func (e *Environment) ApplyEffects(heading float64, velocity models.Velocity) models.Velocity {
	if e == nil || !e.enabled {
		return velocity
	}

	result := velocity

	// Apply wind effect
	if e.wind != nil {
		result = e.wind.Apply(heading, result)
	}

	// Future: Apply other effects like humidity, air density, etc.
	// if e.humidity != nil {
	//     result = applyHumidityEffect(result, *e.humidity)
	// }

	return result
}

// GetState returns environment state for API responses.
func (e *Environment) GetState() *models.EnvironmentState {
	if e == nil || !e.enabled {
		return nil
	}

	state := &models.EnvironmentState{}

	if e.wind != nil {
		state.Wind = e.wind.GetVector()
	}

	if e.humidity != nil {
		state.Humidity = e.humidity
	}

	return state
}

// GetWind returns the wind effect if enabled.
func (e *Environment) GetWind() *WindEffect {
	if e == nil {
		return nil
	}
	return e.wind
}

// IsEnabled returns whether environment effects are enabled.
func (e *Environment) IsEnabled() bool {
	return e != nil && e.enabled
}
