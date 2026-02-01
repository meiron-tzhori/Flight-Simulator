package geo

import (
	"math"
	"testing"
)

func TestHaversine(t *testing.T) {
	tests := []struct {
		name      string
		lat1      float64
		lon1      float64
		lat2      float64
		lon2      float64
		expected  float64 // meters
		tolerance float64 // meters
	}{
		{
			name:      "Same point",
			lat1:      32.0853,
			lon1:      34.7818,
			lat2:      32.0853,
			lon2:      34.7818,
			expected:  0,
			tolerance: 1,
		},
		{
			name:      "Tel Aviv to Jerusalem (~54km actual)",
			lat1:      32.0853,
			lon1:      34.7818,
			lat2:      31.7683,
			lon2:      35.2137,
			expected:  54000,
			tolerance: 2000,
		},
		{
			name:      "Short distance (1 degree latitude ~111km)",
			lat1:      32.0,
			lon1:      34.0,
			lat2:      33.0,
			lon2:      34.0,
			expected:  111000,
			tolerance: 1000,
		},
		{
			name:      "Equator to North Pole",
			lat1:      0,
			lon1:      0,
			lat2:      90,
			lon2:      0,
			expected:  10001965, // ~10,000km
			tolerance: 10000,
		},
		{
			name:      "Across date line",
			lat1:      0,
			lon1:      179,
			lat2:      0,
			lon2:      -179,
			expected:  222390, // ~2 degrees longitude at equator
			tolerance: 5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Haversine(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			diff := math.Abs(result - tt.expected)
			if diff > tt.tolerance {
				t.Errorf("Haversine() = %.2f, expected %.2f ± %.2f (diff: %.2f)",
					result, tt.expected, tt.tolerance, diff)
			}
		})
	}
}

func TestBearing(t *testing.T) {
	tests := []struct {
		name      string
		lat1      float64
		lon1      float64
		lat2      float64
		lon2      float64
		expected  float64 // degrees
		tolerance float64 // degrees
	}{
		{
			name:      "North",
			lat1:      32.0,
			lon1:      34.0,
			lat2:      33.0,
			lon2:      34.0,
			expected:  0,
			tolerance: 1,
		},
		{
			name:      "East",
			lat1:      32.0,
			lon1:      34.0,
			lat2:      32.0,
			lon2:      35.0,
			expected:  90,
			tolerance: 1,
		},
		{
			name:      "South",
			lat1:      32.0,
			lon1:      34.0,
			lat2:      31.0,
			lon2:      34.0,
			expected:  180,
			tolerance: 1,
		},
		{
			name:      "West",
			lat1:      32.0,
			lon1:      34.0,
			lat2:      32.0,
			lon2:      33.0,
			expected:  270,
			tolerance: 1,
		},
		{
			name:      "Northeast (actual ~40°)",
			lat1:      32.0,
			lon1:      34.0,
			lat2:      33.0,
			lon2:      35.0,
			expected:  40,
			tolerance: 2,
		},
		{
			name:      "Tel Aviv to Jerusalem (Southeast)",
			lat1:      32.0853,
			lon1:      34.7818,
			lat2:      31.7683,
			lon2:      35.2137,
			expected:  140,
			tolerance: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bearing(tt.lat1, tt.lon1, tt.lat2, tt.lon2)

			// Normalize both to 0-360
			result = math.Mod(result+360, 360)
			expected := math.Mod(tt.expected+360, 360)

			// Handle wraparound (e.g., 359 and 1 are close)
			diff := math.Abs(result - expected)
			if diff > 180 {
				diff = 360 - diff
			}

			if diff > tt.tolerance {
				t.Errorf("Bearing() = %.2f°, expected %.2f° ± %.2f° (diff: %.2f°)",
					result, expected, tt.tolerance, diff)
			}
		})
	}
}

func TestBearingReversibility(t *testing.T) {
	// Test that bearing from A to B and B to A differ by ~180 degrees
	lat1, lon1 := 32.0, 34.0
	lat2, lon2 := 33.0, 35.0

	bearingAB := Bearing(lat1, lon1, lat2, lon2)
	bearingBA := Bearing(lat2, lon2, lat1, lon1)

	// Normalize both
	bearingAB = math.Mod(bearingAB+360, 360)
	bearingBA = math.Mod(bearingBA+360, 360)

	// Calculate difference
	diff := math.Abs(bearingAB - bearingBA)
	if diff > 180 {
		diff = 360 - diff
	}

	// Should differ by ~180 degrees (allow 10 degree tolerance for non-meridian paths)
	expectedDiff := 180.0
	actualDiff := math.Abs(diff - expectedDiff)

	if actualDiff > 10 {
		t.Errorf("Bearing reversibility: AB=%.2f°, BA=%.2f°, diff=%.2f°, expected ~180° (off by %.2f°)",
			bearingAB, bearingBA, diff, actualDiff)
	}
}

func BenchmarkHaversine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Haversine(32.0853, 34.7818, 31.7683, 35.2137)
	}
}

func BenchmarkBearing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Bearing(32.0853, 34.7818, 31.7683, 35.2137)
	}
}
