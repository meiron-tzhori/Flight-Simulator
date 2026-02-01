package environment

import (
	"math"
	"testing"

	"github.com/meiron-tzhori/Flight-Simulator/internal/models"
)

func TestWindEffect_Apply(t *testing.T) {
	tests := []struct {
		name             string
		windDirection    float64 // degrees (wind FROM direction)
		windSpeed        float64 // m/s
		aircraftHeading  float64 // degrees
		aircraftSpeed    float64 // m/s (airspeed)
		expectedGS       float64 // m/s (ground speed)
		tolerance        float64 // m/s
	}{
		{
			name:            "No wind",
			windDirection:   0,
			windSpeed:       0,
			aircraftHeading: 0,
			aircraftSpeed:   50,
			expectedGS:      50,
			tolerance:       0.1,
		},
		{
			name:            "Direct headwind (flying North, wind from North)",
			windDirection:   0,
			windSpeed:       10,
			aircraftHeading: 0,
			aircraftSpeed:   50,
			expectedGS:      40, // 50 - 10
			tolerance:       0.1,
		},
		{
			name:            "Direct tailwind (flying North, wind from South)",
			windDirection:   180,
			windSpeed:       10,
			aircraftHeading: 0,
			aircraftSpeed:   50,
			expectedGS:      60, // 50 + 10
			tolerance:       0.1,
		},
		{
			name:            "Direct headwind (flying East, wind from East)",
			windDirection:   90,
			windSpeed:       15,
			aircraftHeading: 90,
			aircraftSpeed:   100,
			expectedGS:      85, // 100 - 15
			tolerance:       0.1,
		},
		{
			name:            "Perpendicular crosswind (flying North, wind from East)",
			windDirection:   90,
			windSpeed:       20,
			aircraftHeading: 0,
			aircraftSpeed:   50,
			expectedGS:      53.85, // sqrt(50^2 + 20^2) ≈ 53.85
			tolerance:       0.5,
		},
		{
			name:            "45-degree angle wind",
			windDirection:   45,
			windSpeed:       10,
			aircraftHeading: 0,
			aircraftSpeed:   50,
			expectedGS:      50.14, // Approximately (small crosswind effect)
			tolerance:       1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wind := NewWindEffect(tt.windDirection, tt.windSpeed)
			
			velocity := models.Velocity{
				GroundSpeed:   tt.aircraftSpeed,
				VerticalSpeed: 0,
			}
			
			result := wind.Apply(tt.aircraftHeading, velocity)
			
			diff := math.Abs(result.GroundSpeed - tt.expectedGS)
			if diff > tt.tolerance {
				t.Errorf("Apply() ground speed = %.2f m/s, expected %.2f m/s ± %.2f m/s (diff: %.2f)",
					result.GroundSpeed, tt.expectedGS, tt.tolerance, diff)
			}
			
			// Vertical speed should not change
			if result.VerticalSpeed != velocity.VerticalSpeed {
				t.Errorf("Apply() changed vertical speed: got %.2f, want %.2f",
					result.VerticalSpeed, velocity.VerticalSpeed)
			}
		})
	}
}

func TestWindEffect_CalculateHeadwindComponent(t *testing.T) {
	tests := []struct {
		name          string
		windDirection float64
		windSpeed     float64
		aircraftHdg   float64
		expectedHW    float64 // Positive = headwind, Negative = tailwind
		tolerance     float64
	}{
		{
			name:          "Direct headwind",
			windDirection: 0,
			windSpeed:     10,
			aircraftHdg:   0,
			expectedHW:    10, // Full headwind
			tolerance:     0.1,
		},
		{
			name:          "Direct tailwind",
			windDirection: 180,
			windSpeed:     10,
			aircraftHdg:   0,
			expectedHW:    -10, // Full tailwind
			tolerance:     0.1,
		},
		{
			name:          "Perpendicular wind (no headwind component)",
			windDirection: 90,
			windSpeed:     20,
			aircraftHdg:   0,
			expectedHW:    0, // Pure crosswind
			tolerance:     0.1,
		},
		{
			name:          "45-degree headwind",
			windDirection: 45,
			windSpeed:     20,
			aircraftHdg:   0,
			expectedHW:    14.14, // 20 * cos(45°) ≈ 14.14
			tolerance:     0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wind := NewWindEffect(tt.windDirection, tt.windSpeed)
			result := wind.CalculateHeadwindComponent(tt.aircraftHdg)
			
			diff := math.Abs(result - tt.expectedHW)
			if diff > tt.tolerance {
				t.Errorf("CalculateHeadwindComponent() = %.2f m/s, expected %.2f m/s ± %.2f m/s",
					result, tt.expectedHW, tt.tolerance)
			}
		})
	}
}

func TestWindEffect_CalculateCrosswindComponent(t *testing.T) {
	tests := []struct {
		name          string
		windDirection float64
		windSpeed     float64
		aircraftHdg   float64
		expectedXW    float64 // Positive = from right, Negative = from left
		tolerance     float64
	}{
		{
			name:          "Direct headwind (no crosswind)",
			windDirection: 0,
			windSpeed:     10,
			aircraftHdg:   0,
			expectedXW:    0,
			tolerance:     0.1,
		},
		{
			name:          "Perpendicular wind from right",
			windDirection: 90,
			windSpeed:     20,
			aircraftHdg:   0,
			expectedXW:    20, // Full crosswind from right
			tolerance:     0.1,
		},
		{
			name:          "Perpendicular wind from left",
			windDirection: 270,
			windSpeed:     20,
			aircraftHdg:   0,
			expectedXW:    -20, // Full crosswind from left
			tolerance:     0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wind := NewWindEffect(tt.windDirection, tt.windSpeed)
			result := wind.CalculateCrosswindComponent(tt.aircraftHdg)
			
			diff := math.Abs(result - tt.expectedXW)
			if diff > tt.tolerance {
				t.Errorf("CalculateCrosswindComponent() = %.2f m/s, expected %.2f m/s ± %.2f m/s",
					result, tt.expectedXW, tt.tolerance)
			}
		})
	}
}

func TestWindEffect_GetVector(t *testing.T) {
	wind := NewWindEffect(270.0, 15.5)
	vector := wind.GetVector()
	
	if vector.Direction != 270.0 {
		t.Errorf("GetVector().Direction = %.2f, want 270.0", vector.Direction)
	}
	
	if vector.Speed != 15.5 {
		t.Errorf("GetVector().Speed = %.2f, want 15.5", vector.Speed)
	}
}

func BenchmarkWindEffect_Apply(b *testing.B) {
	wind := NewWindEffect(270, 15)
	velocity := models.Velocity{GroundSpeed: 50, VerticalSpeed: 0}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wind.Apply(45, velocity)
	}
}
