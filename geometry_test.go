package main

import "testing"

func TestCalculateDistance(t *testing.T) {
	// 1. Preparing the input
	// Railway Station
	station := Location{Lat: 60.171, Lon: 24.941}
	// Suomenlinna Island
	island := Location{Lat: 60.144, Lon: 24.984}

	// 2. ACT
	result := CalculateDistance(station, island)

	// 3. ASSERT (Check the result)
	// Google Maps says this is about 3.9 km (3900 meters).
	// Our math might be slightly different depending on exact radius,
	// so let's say: if it's between 3800 and 4000, it's correct.

	expectedMin := int64(3800)
	expectedMax := int64(4000)

	if result < expectedMin || result > expectedMax {
		t.Errorf("Expected distance between 3800m and 4000m, but got %d", result)
	}
}
