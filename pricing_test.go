package main

import (
	"testing"

	"github.com/lnemenl/wolt_1/internal/core/domain"
)

func TestCalculateDeliveryFee(t *testing.T) {
	// 1. PREPARE (The Rules)
	// This is a "Slice Literal". We create the slice and fill it at the same time.
	woltRules := []domain.DistanceRange{
		// 0 to 500m: Free variable fee
		{Min: 0, Max: 500, A: 0, B: 0},
		// 500 to 1000m: 100 cents + (1 * dist / 10)
		{Min: 500, Max: 1000, A: 100, B: 1},
		// 1000m+: Delivery Impossible
		{Min: 1000, Max: 0, A: 0, B: 0},
	}

	// 2. SCENARIO A: Short distance (100m)
	// Should be: A(0) + B(0) = 0 extra fee
	fee, err := CalculateDeliveryFee(100, woltRules)

	t.Logf("Tested Distance: 100m. Result Fee: %d. Error: %v", fee, err)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if fee != 0 {
		t.Errorf("Expected fee 0, got %d", fee)
	}

	// 3. SCENARIO B: Medium distance (600m)
	// Logic: A(100) + (1 * 600 / 10) = 100 + 60 = 160
	fee, err = CalculateDeliveryFee(600, woltRules)

	t.Logf("Tested Distance: 600m. Result Fee: %d. Error: %v", fee, err)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if fee != 160 {
		t.Errorf("Expected fee 160, got %d", fee)
	}

	// 4. SCENARIO C: Too far (1500m)
	// Should return an error
	fee, err = CalculateDeliveryFee(1500, woltRules)

	t.Logf("Tested Distance: 1500m. Result Fee: %d. Error: %v", fee, err)

	if err == nil {
		t.Error("Expected an error for long distance, but got none")
	}
}
