package geo

import "math"

// Bearing calculates the initial bearing from point 1 to point 2.
// Returns bearing in degrees (0-360, where 0 is North).
func Bearing(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert to radians
	lat1Rad := toRadians(lat1)
	lat2Rad := toRadians(lat2)
	dLon := toRadians(lon2 - lon1)

	// Calculate bearing
	y := math.Sin(dLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(dLon)

	bearingRad := math.Atan2(y, x)
	bearingDeg := toDegrees(bearingRad)

	// Normalize to 0-360
	return math.Mod(bearingDeg+360, 360)
}

func toDegrees(radians float64) float64 {
	return radians * 180.0 / math.Pi
}
