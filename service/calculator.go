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

type DeliveryInput struct {
	CartValue int
	UserLat   float64
	UserLon   float64
}

type VenueInfo struct {
	Coordinates [2]float64
	Pricing     models.DeliveryPricing
	MinOrder    int
}

func MergeToVenueInfo(static models.VenueStatic, dynamic models.VenueDynamic) (VenueInfo, error) {
	venueCoords := static.VenueRaw.Location.Coordinates
	if len(venueCoords) < 2 {
		return VenueInfo{}, fmt.Errorf("%w: expected 2 coordinates, got %d", ErrInvalidVenueData, len(venueCoords))
	}
	pricing := dynamic.VenueRaw.DeliverySpecs.DeliveryPricing
	if len(pricing.DistanceRanges) == 0 {
		return VenueInfo{}, fmt.Errorf("%w: missing distance ranges", ErrInvalidVenueData)
	}

	minOrder := dynamic.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge

	return VenueInfo{
		Coordinates: [2]float64{venueCoords[0], venueCoords[1]},
		Pricing:     pricing,
		MinOrder:    minOrder,
	}, nil
}

func CalculatePrice(input DeliveryInput, venue VenueInfo) (models.PriceResponse, error) {
	distance := calculateDistance(input.UserLat, input.UserLon, venue.Coordinates)

	fee, err := calculateFee(distance, venue.Pricing)
	if err != nil {
		return models.PriceResponse{}, err
	}

	surcharge := 0
	if input.CartValue < venue.MinOrder {
		surcharge = venue.MinOrder - input.CartValue
	}

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

func calculateDistance(userLat, userLon float64, venueCoords [2]float64) int {
	venueLon := venueCoords[0]
	venueLat := venueCoords[1]

	// Earth radius in meters
	const R = 6371000.0

	toRad := func(deg float64) float64 {
		return deg * math.Pi / 180
	}

	lat1 := toRad(userLat)
	lat2 := toRad(venueLat)
	dLat := toRad(venueLat - userLat)
	dLon := toRad(venueLon - userLon)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c

	return int(math.Round(distance))
}

func calculateFee(distance int, pricing models.DeliveryPricing) (int, error) {
	for _, r := range pricing.DistanceRanges {
		if distance < r.Min {
			continue
		}

		if r.Max == 0 {
			return 0, fmt.Errorf("%w: %d meters", ErrDistanceTooLong, distance)
		}

		if distance >= r.Max {
			continue
		}

		componentB := int(math.Round(r.B * float64(distance) / 10.0))
		return pricing.BasePrice + r.A + componentB, nil
	}

	return 0, fmt.Errorf("%w: %d meters", ErrNoRangeFound, distance)
}
