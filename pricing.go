package main

import (
	"errors"
	"math"

	"github.com/lnemenl/wolt_1/internal/core/domain"
)

// 2. The Calculation Function
// We take the distance (from our Geometry file) and the list of rules (Slice)
// '[]DistanceRange' -> This means "A slice of DistanceRange objects"
func CalculateDeliveryFee(distance int64, ranges []domain.DistanceRange) (int64, error) {

	// We loop through the slice using 'range'
	// 'i' is the index (0, 1, 2...) which we don't need, so we use '_' to ignore it
	// 'r' is the actual rule (the struct)
	for _, r := range ranges {

		// Check if our distance falls into this rule's bucket
		// Wolt says: "max: 0" means we hit the limit -> delivery impossible
		if r.Max == 0 {
			if distance >= r.Min {
				return 0, errors.New("Delivery not possible for this distance")
			}
			// Edge case: if it's 0 but we are smaller than Min, we just skip
			continue
		}
		if distance >= r.Min && distance < r.Max {

			// FORMULA: base + a + (b * distance / 10)
			// calculate the variable part first.
			// cast to float64 for math, then Round, then cast back to int64.
			variableFee := math.Round(r.B * float64(distance) / 10.0)
			totalFee := r.A + int64(variableFee)
			// nil for error
			return totalFee, nil
		}
	}
	// in case we finish the loop and find nothing
	return 0, errors.New("No matching price rule found")
}
