package algo

import (
	"math"
)

// CalculateVincentyDistance calculates the distance between two points (lat1, lon1) and (lat2, lon2) using Vincenty's formula
func CalculateVincentyDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// WGS-84 ellipsiod parameters
	a := 6378137.0 // semi-major axis in meters
	f := 1 / 298.257223563 // flattening
	b := (1 - f) * a // semi-minor axis

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180.0
	lon1Rad := lon1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	lon2Rad := lon2 * math.Pi / 180.0

	// U1 and U2 are the reduced latitudes (in radians)
	U1 := math.Atan((1 - f) * math.Tan(lat1Rad))
	U2 := math.Atan((1 - f) * math.Tan(lat2Rad))

	// Difference in longitude
	L := lon2Rad - lon1Rad

	// Iterate to solve the Vincenty equation
	sinU1 := math.Sin(U1)
	cosU1 := math.Cos(U1)
	sinU2 := math.Sin(U2)
	cosU2 := math.Cos(U2)

	sinSigma := 0.0
	cosSigma := 0.0
	sigma := 0.0
	sinAlpha := 0.0
	cos2Alpha := 0.0
	 cos2SigmaM := 0.0
	C := 0.0
	lamda := L
	for {
		sinSigma = math.Sqrt(math.Pow(cosU2*math.Sin(lamda), 2) + math.Pow(cosU1*sinU2-math.Sin(U1)*cosU2*math.Cos(lamda), 2))
		cosSigma = sinU1*sinU2 + cosU1*cosU2*math.Cos(lamda)
		sigma = math.Atan2(sinSigma, cosSigma)

		sinAlpha = cosU1 * cosU2 * math.Sin(lamda) / sinSigma
		cos2Alpha = 1 - sinAlpha*sinAlpha
		cos2SigmaM = cosSigma - 2*sinU1*sinU2/cos2Alpha
		C = f / 16 * cos2Alpha * (4 + f * (4 - 3*cos2Alpha))

		// Update lambda (the difference in longitude)
		lamdaPrev := lamda
		lamda = L + (1 - C) * f * sinAlpha *
			(sigma + C*sinSigma*(cos2SigmaM + C*cosSigma*(-1 + 2*cos2SigmaM*cos2SigmaM)))

		if math.Abs(lamda-lamdaPrev) < 1e-12 {
			break
		}
	}

	// Calculate distance
	u2 := cos2Alpha * (a*a - b*b) / (b * b)
	A := 1 + u2/16384*(4096+u2*(-768+u2*(320-175*u2)))
	B := u2 / 1024 * (256+u2*(-128+u2*(74-47*u2)))
	distance := b * A * (sigma - B*sinSigma*(cos2SigmaM + B*cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)))
	
	// Return distance in kilometers
	return distance / 1000.0 // Convert from meters to kilometers
}
