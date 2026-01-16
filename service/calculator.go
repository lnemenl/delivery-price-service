package service

import (
	"errors"
	"math"

	"github.com/lnemenl/wolt_1/models"
)

// DeliveryInput groups the user-provided data
type DeliveryInput struct {
	CartValue int
	UserLat   float64
	UserLon   float64
}

// CalculatePrice computes the total price breakdown
func CalculatePrice(input DeliveryInput, static models.VenueStatic, dynamic models.VenueDynamic) (models.PriceResponse, error) {
	// 1. Calculate Distance
	venueCoords := static.VenueRaw.Location.Coordinates
	// Safety check: Ensure API gave us both Lat and Lon
	if len(venueCoords) < 2 {
		return models.PriceResponse{}, errors.New("venue location data is incomplete")
	}
	distance := calculateDistance(input.UserLat, input.UserLon, venueCoords)

	// 2. Calculate Delivery Fee
	pricing := dynamic.VenueRaw.DeliverySpecs.DeliveryPricing
	fee, err := calculateFee(distance, pricing)
	if err != nil {
		return models.PriceResponse{}, err
	}

	// 3. Calculate Surcharge
	// Surcharge is the difference if cart value is below minimum
	surcharge := 0
	minOrder := dynamic.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge
	if input.CartValue < minOrder {
		surcharge = minOrder - input.CartValue
	}

	// 4. Calculate Total
	total := input.CartValue + surcharge + fee

	return models.PriceResponse{
		TotalPrice:          total,
		SmallOrderSurcharge: surcharge,
		CartValue:           input.CartValue,
		Delivery: models.DeliveryDetails{
			Fee:      fee,
			Distance: distance,
		},
	}, nil
}

// // calculateDistance returns straight-line distance in meters
// func calculateDistance(userLat, userLon float64, venueCoords []float64) int {
// 	venueLon := venueCoords[0]
// 	venueLat := venueCoords[1]

// 	// Conversion constant: 1 degree latitude ~= 111,139 meters
// 	const metersPerDegree = 111139.0

// 	latDist := (userLat - venueLat) * metersPerDegree
// 	lonDist := (userLon - venueLon) * metersPerDegree

// 	// Pythagoras: c = sqrt(a^2 + b^2)
// 	distInMeters := math.Sqrt(latDist*latDist + lonDist*lonDist)

// 	return int(math.Round(distInMeters))
// }

// the Haversine formula
func calculateDistance(userLat, userLon float64, venueCoords []float64) int {
	venueLon := venueCoords[0]
	venueLat := venueCoords[1]

	// Earth radius in meters
	const R = 6371000.0

	// Convert degrees to radians
	toRad := func(deg float64) float64 {
		return deg * math.Pi / 180
	}

	lat1 := toRad(userLat)
	lat2 := toRad(venueLat)
	dLat := toRad(venueLat - userLat)
	dLon := toRad(venueLon - userLon)

	// Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c

	return int(math.Round(distance))
}

// calculateFee finds the matching price rule for the distance
func calculateFee(distance int, pricing models.DeliveryPricing) (int, error) {
	for _, r := range pricing.DistanceRanges {
		// Logic:
		// 1. Check if distance is above the minimum (inclusive)
		// 2. Check if distance is below the maximum (exclusive)
		// 3. EXCEPTION: The README says: "max: 0 means delivery is NOT available for distances >= min".

		if distance >= r.Min {
			// If Max is 0, we hit the limit. Delivery is impossible
			if r.Max == 0 {
				return 0, errors.New("delivery distance too long")
			}

			// If distance is within range, calculate price.
			if distance < r.Max {
				// Formula: Base + A + (B * distance / 10)
				// We convert distance to float to match B, multiply them, then round the result back to int
				componentB := int(math.Round(r.B * float64(distance) / 10.0))
				return pricing.BasePrice + r.A + componentB, nil
			}
		}
	}

	return 0, errors.New("no matching delivery range found")
}
