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

func CalculatePrice(input DeliveryInput, static models.VenueStatic, dynamic models.VenueDynamic) (models.PriceResponse, error) {
	venueCoords := static.VenueRaw.Location.Coordinates
	if len(venueCoords) < 2 {
		return models.PriceResponse{}, fmt.Errorf("%w: expected 2 coordinates, got %d", ErrInvalidVenueData, len(venueCoords))
	}
	distance := calculateDistance(input.UserLat, input.UserLon, [2]float64{venueCoords[0], venueCoords[1]})

	pricing := dynamic.VenueRaw.DeliverySpecs.DeliveryPricing
	fee, err := calculateFee(distance, pricing)
	if err != nil {
		return models.PriceResponse{}, err
	}

	surcharge := 0
	minOrder := dynamic.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge
	if input.CartValue < minOrder {
		surcharge = minOrder - input.CartValue
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
