package service_test

import (
	"testing"

	"github.com/lnemenl/wolt_1/internal/core/domain"
	"github.com/lnemenl/wolt_1/internal/core/service"
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
	fee, err := service.CalculateDeliveryFee(100, woltRules)

	t.Logf("Tested Distance: 100m. Result Fee: %d. Error: %v", fee, err)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if fee != 0 {
		t.Errorf("Expected fee 0, got %d", fee)
	}

	// 3. SCENARIO B: Medium distance (600m)
	// Logic: A(100) + (1 * 600 / 10) = 100 + 60 = 160
	fee, err = service.CalculateDeliveryFee(600, woltRules)

	t.Logf("Tested Distance: 600m. Result Fee: %d. Error: %v", fee, err)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if fee != 160 {
		t.Errorf("Expected fee 160, got %d", fee)
	}

	// 4. SCENARIO C: Too far (1500m)
	// Should return an error
	fee, err = service.CalculateDeliveryFee(1500, woltRules)

	t.Logf("Tested Distance: 1500m. Result Fee: %d. Error: %v", fee, err)

	if err == nil {
		t.Error("Expected an error for long distance, but got none")
	}
}

func TestCalculateSmallOrderSurcharge(t *testing.T) {
	// 1. ARRANGE (The Scenarios)
	// We use a "Table-Driven Test". This is the Go standard
	// Instead of writing 4 different "if" blocks, we make a list of inputs and expected outputs
	tests := []struct {
		name      string // Name of the scenario (for logs)
		cartValue int64  // Input: How much the user bought
		minValue  int64  // Input: The minimum order rule (e.g. 1000)
		expected  int64  // Output: What the surcharge should be
	}{
		// Case 1: Simple. 800 is less than 1000. Difference is 200.
		{"Below Minimum", 800, 1000, 200},

		// Case 2: Exact. 1000 equals 1000. No surcharge.
		{"Exact Minimum", 1000, 1000, 0},

		// Case 3: Above. 1500 is more than 1000. No surcharge.
		{"Above Minimum", 1500, 1000, 0},

		// Case 4: Zero. Bought nothing. Pay full surcharge.
		{"Zero Cart", 0, 1000, 1000},
	}

	// 2. ACT & ASSERT (The Loop)
	for _, tt := range tests {
		// t.Run creates a sub-test. If one fails, we know exactly which name it was
		t.Run(tt.name, func(t *testing.T) {

			// We call the function (which doesn't exist yet!)
			got := service.CalculateSmallOrderSurcharge(tt.cartValue, tt.minValue)

			// We check the result
			if got != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, got)
			}
		})
	}
}
