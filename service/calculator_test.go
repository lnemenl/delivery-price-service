package service

import (
	"testing"

	"github.com/lnemenl/wolt_1/models"
)

func TestCalculatePrice(t *testing.T) {

	// =========================================================================
	// SETTING THE SCENE
	// =========================================================================

	// 1. The Venue is at (0.0, 0.0)
	// We use 0,0 because calculating distance from zero is easy to verify
	venueLoc := models.VenueStatic{}
	venueLoc.VenueRaw.Location.Coordinates = []float64{0.0, 0.0}

	// 2. The Rules of the Game
	// These rules mimic a real response from the Wolt API
	// - Base Price:    1.90€ (190 cents)
	// - Minimum Order: 10.00€ (1000 cents)
	// - Range 1 (0-500m):    Free distance fee. (Price = Base + 0 + 0)
	// - Range 2 (500-1000m): Expensive (Price = Base + 100 + (5 * distance / 10))
	// - Range 3 (1000m+):    Impossible (Max=0 means "We do not deliver here")
	venueRules := models.VenueDynamic{}
	venueRules.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge = 1000
	venueRules.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice = 190
	venueRules.VenueRaw.DeliverySpecs.DeliveryPricing.DistanceRanges = []models.DistanceRange{
		{Min: 0, Max: 500, A: 0, B: 0},
		{Min: 500, Max: 1000, A: 100, B: 5.0},
		{Min: 1000, Max: 0, A: 0, B: 0},
	}

	// =========================================================================
	// CHAPTER 1: THE HAPPY CUSTOMER (Range 1)
	// Scenario: User is right next to the restaurant (0m). Order is 10€
	// Expectation: Only Base Price. No Surcharge
	// =========================================================================
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

		// 1. Surcharge Check (Cart 1000 >= Min 1000) -> 0
		if resp.SmallOrderSurcharge != 0 {
			t.Errorf("Expected 0 surcharge, got %d", resp.SmallOrderSurcharge)
		}

		// 2. Fee Check (Range 1: Base 190 + A 0 + B 0) -> 190
		if resp.Delivery.Fee != 190 {
			t.Errorf("Expected 190 fee, got %d", resp.Delivery.Fee)
		}

		// 3. Total Check (1000 + 0 + 190) -> 1190
		if resp.TotalPrice != 1190 {
			t.Errorf("Expected total 1190, got %d", resp.TotalPrice)
		}
	})

	// =========================================================================
	// CHAPTER 2: THE SMALL ORDER SURCHARGE
	// Scenario: User is close, but only buys a healthy salad (8.00€)
	// Expectation: A 2.00€ surcharge is added to fill the gap to 10.00€
	// =========================================================================
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

		// Logic: Surcharge = Min (1000) - Cart (800) = 200
		if resp.SmallOrderSurcharge != 200 {
			t.Errorf("Expected surcharge 200, got %d", resp.SmallOrderSurcharge)
		}

		// Logic: Total = 800 + 200 + 190 = 1190
		if resp.TotalPrice != 1190 {
			t.Errorf("Expected total 1190, got %d", resp.TotalPrice)
		}
	})

	// =========================================================================
	// CHAPTER 3: THE EXPENSIVE ZONE
	// Scenario: User is 667m away. This falls into Range 2 (500-1000m)
	// Formula: Base + A + (B * distance / 10)
	// Values:  190  + 100 + (5.0 * 667 / 10)
	// =========================================================================
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

		// Verify Distance
		if resp.Delivery.Distance != 667 {
			t.Errorf("Expected distance 667m, got %d", resp.Delivery.Distance)
		}

		// Verify Fee
		// Math: (5.0 * 667) / 10  = 3335 / 10 = 333.5
		// Round(333.5) = 334
		// Fee = Base(190) + A(100) + B_Component(334) = 624
		if resp.Delivery.Fee != 624 {
			t.Errorf("Expected fee 624, got %d", resp.Delivery.Fee)
		}
	})

	// =========================================================================
	// CHAPTER 4: NEGATIVE COORDINATES
	// Scenario: User is at -0.006 Latitude. Distance should still be positive 667m
	// This proves Math.Sqrt() logic works for absolute distances
	// =========================================================================
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

		// Distance MUST be positive 667, NOT -667
		if resp.Delivery.Distance != 667 {
			t.Errorf("Expected distance 667m, got %d", resp.Delivery.Distance)
		}
	})

	// =========================================================================
	// CHAPTER 5: THE FORBIDDEN ZONE
	// Scenario: User is 1111m away. Range 3 starts at 1000m and has Max=0
	// Expectation: The calculator returns an error
	// =========================================================================
	t.Run("Edge Case: Distance Too Far", func(t *testing.T) {
		input := DeliveryInput{
			CartValue: 1000,
			UserLat:   0.010, // 1111 meters
			UserLon:   0.0,
		}

		_, err := CalculatePrice(input, venueLoc, venueRules)

		// We EXPECT an error here
		if err == nil {
			t.Fatal("Expected an error (Too Far), but got success!")
		}
	})

	// =========================================================================
	// CHAPTER 6: ZERO CART VALUE
	// Scenario: User orders 0€ worth of items
	// Expectation: Surcharge is 1000 (full minimum), fee is calculated normally
	// =========================================================================
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

		// Surcharge should be full minimum (1000)
		if resp.SmallOrderSurcharge != 1000 {
			t.Errorf("Expected surcharge 1000, got %d", resp.SmallOrderSurcharge)
		}

		// Total should be 0 + 1000 + 190 = 1190
		if resp.TotalPrice != 1190 {
			t.Errorf("Expected total 1190, got %d", resp.TotalPrice)
		}
	})

	// =========================================================================
	// CHAPTER 7: NEGATIVE B COEFFICIENT
	// Scenario: B can be negative (example from real data: b = -1)
	// Formula: Base + A + (B * distance / 10)
	// If B is negative, it reduces the fee
	// =========================================================================
	t.Run("Edge Case: Negative B Coefficient", func(t *testing.T) {
		input := DeliveryInput{
			CartValue: 1000,
			UserLat:   0.006,
			UserLon:   0.0,
		}

		// Custom rules with negative B
		customRules := models.VenueDynamic{}
		customRules.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge = 1000
		customRules.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice = 190
		customRules.VenueRaw.DeliverySpecs.DeliveryPricing.DistanceRanges = []models.DistanceRange{
			{Min: 0, Max: 1000, A: 1000, B: -1.0}, // Negative multiplier
			{Min: 1000, Max: 0, A: 0, B: 0},
		}

		resp, err := CalculatePrice(input, venueLoc, customRules)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Distance is 667m
		// Fee = Base(190) + A(1000) + B_Component(-1.0 * 667 / 10)
		// B_Component = int(Round(-66.7)) = -67
		// Fee = 190 + 1000 + (-67) = 1123
		if resp.Delivery.Fee != 1123 {
			t.Errorf("Expected fee 1123, got %d", resp.Delivery.Fee)
		}
	})
}
