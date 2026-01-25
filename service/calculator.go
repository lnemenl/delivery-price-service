package service

import (
	"errors"
	"fmt"
	"math"

	"github.com/lnemenl/delivery-price-service/models"
)

var ErrDistanceTooLong = errors.New("delivery distance too long")
var ErrNoRangeFound = errors.New("no matching delivery range found")
var ErrInvalidVenueData = errors.New("venue location data is incomplete")

// DeliveryInput groups the user-provided data
type DeliveryInput struct {
	CartValue int
	UserLat   float64
	UserLon   float64
}

// CalculatePrice computes the total price breakdown
func CalculatePrice(input DeliveryInput, static models.VenueStatic, dynamic models.VenueDynamic) (models.PriceResponse, error) {
	// Extract venue coordinates
	venueCoords := static.VenueRaw.Location.Coordinates
	// Verify venue location data contains both latitude and longitude
	if len(venueCoords) < 2 {
		return models.PriceResponse{}, fmt.Errorf("%w: expected 2 coordinates, got %d", ErrInvalidVenueData, len(venueCoords))
	}
	distance := calculateDistance(input.UserLat, input.UserLon, [2]float64{venueCoords[0], venueCoords[1]})

	// Calculate delivery fee based on distance and pricing rules
	pricing := dynamic.VenueRaw.DeliverySpecs.DeliveryPricing
	fee, err := calculateFee(distance, pricing)
	if err != nil {
		return models.PriceResponse{}, err
	}

	// Calculate surcharge as the difference if cart value is below minimum
	surcharge := 0
	minOrder := dynamic.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge
	if input.CartValue < minOrder {
		surcharge = minOrder - input.CartValue
	}

	// Calculate total price
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

// Use Haversine formula to calculate accurate distance on Earth surface
func calculateDistance(userLat, userLon float64, venueCoords [2]float64) int {
	venueLon := venueCoords[0]
	venueLat := venueCoords[1]

	// Earth radius in meters
	const R = 6371000.0

	// Convert degrees to radians for trigonometric calculations
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

// Find the matching distance range and calculate delivery fee
func calculateFee(distance int, pricing models.DeliveryPricing) (int, error) {
	for _, r := range pricing.DistanceRanges {
		// Check if distance falls within current range:
		// Range is [min, max) - minimum inclusive, maximum exclusive
		// Exception: max=0 means delivery unavailable for distances >= min

		if distance >= r.Min {
			// Check if delivery is available (max=0 means distance limit reached)
			if r.Max == 0 {
				// Wrap error
				return 0, fmt.Errorf("%w: %d meters", ErrDistanceTooLong, distance)
			}

			// Calculate fee for distance within range
			if distance < r.Max {
				// Fee formula: base_price + a + round(b * distance / 10)
				componentB := int(math.Round(r.B * float64(distance) / 10.0))
				return pricing.BasePrice + r.A + componentB, nil
			}
		}
	}

	return 0, fmt.Errorf("%w: %d meters", ErrNoRangeFound, distance)
}
