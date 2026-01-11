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

// CalculatePrice acts as the "Manager". It delegates work to helper functions
func CalculatePrice(input DeliveryInput, static models.VenueStatic, dynamic models.VenueDynamic) (models.PriceResponse, error) {

	// Step 1: Get the distance from the Location Specialist
	venueCoords := static.VenueRaw.Location.Coordinates
	distanceMeters := calculateDistance(input.UserLat, input.UserLon, venueCoords)

	// Step 2: Get the fee from the Pricing Specialist
	pricingRules := dynamic.VenueRaw.DeliverySpecs.DeliveryPricing
	deliveryFee, err := calculateDeliveryFee(distanceMeters, pricingRules)
	if err != nil {
		return models.PriceResponse{}, err
	}

	// Step 3: Get the surcharge from the Surcharge Specialist
	minOrderPrice := dynamic.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge
	surcharge := calculateSurcharge(input.CartValue, minOrderPrice)

	// Step 4: Sum it up
	totalPrice := input.CartValue + surcharge + deliveryFee

	// Step 5: Package the result
	return models.PriceResponse{
		TotalPrice:          totalPrice,
		SmallOrderSurcharge: surcharge,
		CartValue:           input.CartValue,
		Delivery: models.DeliveryDetail{
			Fee:      deliveryFee,
			Distance: distanceMeters,
		},
	}, nil
}

// calculateDistance handles the geometry.
// It converts GPS coordinates (degrees) into meters.
func calculateDistance(userLat, userLon float64, venueCoords []float64) int {
	// API gives coordinates as [Longitude, Latitude]
	venueLon := venueCoords[0]
	venueLat := venueCoords[1]

	const metersPerDegree = 111139.0 // Earth's conversion constant

	// Calculate the "gap" in degrees
	latGap := userLat - venueLat
	lonGap := userLon - venueLon

	// Convert gaps to meters
	latMeters := latGap * metersPerDegree
	lonMeters := lonGap * metersPerDegree

	// Pythagoras: a^2 + b^2 = c^2
	distSquared := (latMeters * latMeters) + (lonMeters * lonMeters)
	distMeters := math.Sqrt(distSquared)

	return int(math.Round(distMeters))
}

// calculateDeliveryFee handles the complex pricing rules.
func calculateDeliveryFee(distance int, pricing models.DeliveryPricing) (int, error) {
	for _, rangeRule := range pricing.DistanceRanges {
		// Check if our distance falls inside this rule's range
		// rangeRule.Min is inclusive (>=)
		// rangeRule.Max is exclusive (<), UNLESS it is 0 (which means "infinity")

		isAboveMin := distance >= rangeRule.Min
		isBelowMax := (rangeRule.Max == 0) || (distance < rangeRule.Max)

		if isAboveMin && isBelowMax {
			// Found the matching rule! Calculate the cost.
			// Formula: Base Price + A + (B * Distance / 10)

			distanceComponent := rangeRule.B * float64(distance) / 10.0
			roundedDistanceComponent := int(math.Round(distanceComponent))

			totalFee := pricing.BasePrice + rangeRule.A + roundedDistanceComponent
			return totalFee, nil
		}
	}

	return 0, errors.New("delivery not possible for this distance")
}

// calculateSurcharge handles the small order penalty logic.
func calculateSurcharge(cartValue, minPrice int) int {
	if cartValue < minPrice {
		return minPrice - cartValue
	}
	return 0
}
