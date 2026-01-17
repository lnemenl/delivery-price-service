package service

import (
	"testing"

	"github.com/lnemenl/wolt_1/models"
)

func TestCalculatePrice(t *testing.T) {

	// Set up test data

	// Venue location at origin (0.0, 0.0) for easy distance verification
	venueLoc := models.VenueStatic{}
	venueLoc.VenueRaw.Location.Coordinates = []float64{0.0, 0.0}

	// Pricing configuration mimicking real Wolt API response
	// Base price: 190 cents, Minimum order: 1000 cents
	// Range 1 (0-500m): Price = 190 + 0 + 0 = 190
	// Range 2 (500-1000m): Price = 190 + 100 + (5 * distance / 10)
	// Range 3 (1000m+): Delivery unavailable (max=0)
	venueRules := models.VenueDynamic{}
	venueRules.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge = 1000
	venueRules.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice = 190
	venueRules.VenueRaw.DeliverySpecs.DeliveryPricing.DistanceRanges = []models.DistanceRange{
		{Min: 0, Max: 500, A: 0, B: 0},
		{Min: 500, Max: 1000, A: 100, B: 5.0},
		{Min: 1000, Max: 0, A: 0, B: 0},
	}

	// Test: User at venue location with large order
	// Distance = 0m, Cart = 1000 cents (meets minimum)
	// Expected: No surcharge, base delivery fee only
	t.Run("Happy Path: Close distance, large order", func(t *testing.T) {
		input := DeliveryInput{
			CartValue: 1000,
			UserLat:   0.0,
			UserLon:   0.0,
		}

		resp, err := CalculatePrice(input, venueLoc, venueRules)

		if err != nil {
			t.Fatalf("Expected success, but got error: %v", err)
		}

		// Verify surcharge is 0 when cart value meets minimum
		if resp.SmallOrderSurcharge != 0 {
			t.Errorf("Expected 0 surcharge, got %d", resp.SmallOrderSurcharge)
		}

		// Verify fee = base_price (190) when in range 1
		if resp.Delivery.Fee != 190 {
			t.Errorf("Expected 190 fee, got %d", resp.Delivery.Fee)
		}

		// Verify total = cart (1000) + surcharge (0) + fee (190) = 1190
		if resp.TotalPrice != 1190 {
			t.Errorf("Expected total 1190, got %d", resp.TotalPrice)
		}
	})

	// Test: User at venue with small order below minimum
	// Distance = 0m, Cart = 800 cents (below 1000 minimum)
	// Expected: Surcharge = 1000 - 800 = 200 cents
	t.Run("Logic Check: Small Order Surcharge", func(t *testing.T) {
		input := DeliveryInput{
			CartValue: 800, // 8.00€
			UserLat:   0.0,
			UserLon:   0.0,
		}

		resp, err := CalculatePrice(input, venueLoc, venueRules)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify surcharge = minimum (1000) - cart (800) = 200
		if resp.SmallOrderSurcharge != 200 {
			t.Errorf("Expected surcharge 200, got %d", resp.SmallOrderSurcharge)
		}

		// Verify total = cart (800) + surcharge (200) + fee (190) = 1190
		if resp.TotalPrice != 1190 {
			t.Errorf("Expected total 1190, got %d", resp.TotalPrice)
		}
	})

	// Test: User 667m away in mid-distance range
	// Distance = 667m (in range 2: 500-1000m)
	// Fee = base (190) + a (100) + b*dist/10 (5.0 * 667 / 10 = 334)
	t.Run("Math Check: Complex Calculation with Multiplier B", func(t *testing.T) {
		input := DeliveryInput{
			CartValue: 1000,
			UserLat:   0.006, // 0.006 degrees * 111139 = 667 meters
			UserLon:   0.0,
		}

		resp, err := CalculatePrice(input, venueLoc, venueRules)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify calculated distance
		if resp.Delivery.Distance != 667 {
			t.Errorf("Expected distance 667m, got %d", resp.Delivery.Distance)
		}

		// Verify delivery fee calculation
		// Fee = base (190) + a (100) + round(5.0 * 667 / 10) = 190 + 100 + 334 = 624
		if resp.Delivery.Fee != 624 {
			t.Errorf("Expected fee 624, got %d", resp.Delivery.Fee)
		}
	})

	// Test: Negative coordinates produce correct positive distance
	// User at -0.006 latitude, distance should be 667m (not -667m)
	// Verifies absolute value handling in distance calculation
	t.Run("Math Check: Negative Coordinates", func(t *testing.T) {
		input := DeliveryInput{
			CartValue: 1000,
			UserLat:   -0.006, // Negative!
			UserLon:   0.0,
		}

		resp, err := CalculatePrice(input, venueLoc, venueRules)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify distance is positive 667m
		if resp.Delivery.Distance != 667 {
			t.Errorf("Expected distance 667m, got %d", resp.Delivery.Distance)
		}
	})

	// Test: Distance exceeds delivery limit
	// Distance = 1111m, exceeds range 3 limit (max=0 at min=1000m)
	// Expected: Error returned, delivery unavailable
	t.Run("Edge Case: Distance Too Far", func(t *testing.T) {
		input := DeliveryInput{
			CartValue: 1000,
			UserLat:   0.010, // 1111 meters
			UserLon:   0.0,
		}

		_, err := CalculatePrice(input, venueLoc, venueRules)

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected distance too far error, but got success")
		}
	})

	// Test: Zero cart value applies full minimum surcharge
	// Cart = 0 cents, Minimum = 1000 cents
	// Expected: Surcharge = 1000 cents (full minimum)
	t.Run("Edge Case: Zero Cart Value", func(t *testing.T) {
		input := DeliveryInput{
			CartValue: 0,
			UserLat:   0.0,
			UserLon:   0.0,
		}

		resp, err := CalculatePrice(input, venueLoc, venueRules)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify surcharge equals full minimum when cart is zero
		if resp.SmallOrderSurcharge != 1000 {
			t.Errorf("Expected surcharge 1000, got %d", resp.SmallOrderSurcharge)
		}

		// Verify total = cart (0) + surcharge (1000) + fee (190) = 1190
		if resp.TotalPrice != 1190 {
			t.Errorf("Expected total 1190, got %d", resp.TotalPrice)
		}
	})

	// Test: Negative B coefficient reduces delivery fee
	// B can be negative in real pricing (e.g., promotional discounts)
	// Fee = base (190) + a (1000) + (b * distance / 10)
	t.Run("Edge Case: Negative B Coefficient", func(t *testing.T) {
		input := DeliveryInput{
			CartValue: 1000,
			UserLat:   0.006,
			UserLon:   0.0,
		}

		// Configure pricing with negative B coefficient
		customRules := models.VenueDynamic{}
		customRules.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge = 1000
		customRules.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice = 190
		customRules.VenueRaw.DeliverySpecs.DeliveryPricing.DistanceRanges = []models.DistanceRange{
			{Min: 0, Max: 1000, A: 1000, B: -1.0},
			{Min: 1000, Max: 0, A: 0, B: 0},
		}

		resp, err := CalculatePrice(input, venueLoc, customRules)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Distance = 667m
		// Fee = base (190) + a (1000) + round(-1.0 * 667 / 10)
		// Fee = 190 + 1000 + round(-66.7) = 190 + 1000 - 67 = 1123
		if resp.Delivery.Fee != 1123 {
			t.Errorf("Expected fee 1123, got %d", resp.Delivery.Fee)
		}
	})
}
