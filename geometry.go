package main

import "math"

type Location struct {
	Lat float64 // Latitude
	Lon float64 // Longitude
}

// Function to calculate distance
func CalculateDistance(p1 Location, p2 Location) int64 {
	// 1. Convert degrees to radians (Math requires radians)
	// We multiply by Pi and divide by 180.
	lat1 := p1.Lat * math.Pi / 180
	lon1 := p1.Lon * math.Pi / 180
	lat2 := p2.Lat * math.Pi / 180
	lon2 := p2.Lon * math.Pi / 180

	// 2. The Haversine Formula. It calculates distance on a sphere (Earth)
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Earth radius is 6,371,000 meters
	distance := 6371000 * c

	// We return the distance as an integer (no decimals needed for meters)
	return int64(distance)
}
