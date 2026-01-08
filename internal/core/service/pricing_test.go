package service

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

func TestCalculateSmallOrderSurcharge(t *testing.T) {
	// ARRANGE
	// We define the rule: Minimum order is 1000 cents (10 EUR)
	minimumNoSurcharge := int64(1000)

	// We define test cases (Table Driven Tests are elegant!)
	tests := []struct {
		name      string
		cartValue int64
		expected  int64
	}{
		{"Cart is exactly min", 1000, 0},
		{"Cart is above min", 1500, 0},
		{"Cart is below min", 800, 200}, // 1000 - 800 = 200
		{"Cart is zero", 0, 1000},       // 1000 - 0 = 1000
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ACT
			surcharge := CalculateSmallOrderSurcharge(tt.cartValue, minimumNoSurcharge)

			// ASSERT
			if surcharge != tt.expected {
				t.Errorf("Cart %d: expected surcharge %d, got %d", tt.cartValue, tt.expected, surcharge)
			}
		})
	}
}
