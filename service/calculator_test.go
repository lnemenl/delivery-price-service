package service

import (
	"testing"

	"github.com/lnemenl/wolt_1/models"
)

func TestCalculateDeliveryFee(t *testing.T) {
	// 1. SETUP: We define the "Shape" of a test case
	type testCase struct {
		name        string                 // Name of the scenario
		distance    int                    // Input
		pricing     models.DeliveryPricing // Input
		expectedFee int                    // Expected Output
		expectError bool                   // Expected Failure?
	}

	// 2. DATA: We list our scenarios (The Playlist)
	tests := []testCase{
		{
			name:     "Base fee only (Short distance)",
			distance: 100,
			pricing: models.DeliveryPricing{
				BasePrice: 190,
				DistanceRanges: []models.DistanceRange{
					{Min: 0, Max: 500, A: 0, B: 0},
				},
			},
			expectedFee: 190,
			expectError: false,
		},
		{
			name:     "Fee with distance multiplier",
			distance: 600,
			pricing: models.DeliveryPricing{
				BasePrice: 190,
				DistanceRanges: []models.DistanceRange{
					{Min: 0, Max: 500, A: 0, B: 0},
					{Min: 500, Max: 1000, A: 100, B: 1},
				},
			},
			// Calculation: Base(190) + A(100) + (B(1) * 600m / 10) = 290 + 60 = 350
			expectedFee: 350,
			expectError: false,
		},
		{
			name:     "Too far (No matching range)",
			distance: 1500,
			pricing: models.DeliveryPricing{
				BasePrice: 190,
				DistanceRanges: []models.DistanceRange{
					{Min: 0, Max: 1000, A: 0, B: 0},
				},
			},
			expectedFee: 0,
			expectError: true,
		},
	}

	// 3. EXECUTION: Loop through the playlist
	for _, currentCase := range tests {

		t.Run(currentCase.name, func(t *testing.T) {
			// Log inputs so we can debug easily
			t.Logf("Testing: %s | Distance: %d", currentCase.name, currentCase.distance)

			// CALL THE FUNCTION
			gotFee, err := calculateDeliveryFee(currentCase.distance, currentCase.pricing)

			// CHECK FOR ERROR
			if currentCase.expectError {
				if err == nil {
					t.Fatalf("Expected an error but got none!")
				} else {
					t.Logf("Got expected error: %v", err)
				}
				return // Stop this test case here if we expected an error
			}

			// CHECK IF UNEXPECTED ERROR
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// CHECK THE MATH
			if gotFee != currentCase.expectedFee {
				t.Errorf("Math Wrong! Expected %d, but got %d", currentCase.expectedFee, gotFee)
			} else {
				t.Logf("Success! Fee matches: %d", gotFee)
			}
		})
	}
}
