package service

import (
	"errors"
	"testing"

	"github.com/lnemenl/delivery-price-service/models"
)

func TestCalculatePrice(t *testing.T) {

	standardVenue := VenueInfo{
		Coordinates: [2]float64{0.0, 0.0},
		MinOrder:    1000,
		Pricing: models.DeliveryPricing{
			BasePrice: 190,
			DistanceRanges: []models.DistanceRange{
				{Min: 0, Max: 500, A: 0, B: 0},
				{Min: 500, Max: 1000, A: 100, B: 5.0}, // +1.00 EUR fixed + 0.50 EUR/10m
				{Min: 1000, Max: 0, A: 0, B: 0},       // Delivery not allowed > 1000m
			},
		},
	}

	tests := []struct {
		name          string
		input         DeliveryInput
		venue         VenueInfo
		wantTotal     int
		wantFee       int
		wantSurcharge int
		wantErr       error
	}{
		{
			name: "Happy Path: Close distance, large order",
			input: DeliveryInput{
				CartValue: 1000,
				UserLat:   0.0,
				UserLon:   0.0,
			},
			venue:     standardVenue,
			wantTotal: 1190, // 1000 cart + 190 fee + 0 surcharge
			wantFee:   190,
			wantErr:   nil,
		},
		{
			name: "Logic Check: Small Order Surcharge",
			input: DeliveryInput{
				CartValue: 800, // 8.00 EUR (Below 10.00 min)
				UserLat:   0.0,
				UserLon:   0.0,
			},
			venue:         standardVenue,
			wantTotal:     1190, // 800 cart + 190 fee + 200 surcharge
			wantFee:       190,
			wantSurcharge: 200, // 1000 - 800
			wantErr:       nil,
		},
		{
			name: "Math Check: Complex Calculation (Range 2)",
			input: DeliveryInput{
				CartValue: 1000,
				UserLat:   0.006, // approx 667 meters away
				UserLon:   0.0,
			},
			venue:     standardVenue,
			wantTotal: 1624, // 1000 cart + 624 fee
			wantFee:   624,  // Base(190) + A(100) + B(5.0 * 667 / 10 = 333.5->334) = 624
			wantErr:   nil,
		},
		{
			name: "Math Check: Negative Coordinates",
			input: DeliveryInput{
				CartValue: 1000,
				UserLat:   -0.006, // approx 667 meters away
				UserLon:   0.0,
			},
			venue:     standardVenue,
			wantTotal: 1624,
			wantFee:   624,
			wantErr:   nil,
		},
		{
			name: "Edge Case: Distance Too Far (>1000m)",
			input: DeliveryInput{
				CartValue: 1000,
				UserLat:   0.010, // approx 1111 meters away
				UserLon:   0.0,
			},
			venue:   standardVenue,
			wantErr: ErrDistanceTooLong,
		},
		{
			name: "Edge Case: No Range Defined",
			input: DeliveryInput{
				CartValue: 1000,
				UserLat:   0.020, // approx 2222 meters away
				UserLon:   0.0,
			},
			venue: VenueInfo{
				Coordinates: [2]float64{0.0, 0.0},
				Pricing: models.DeliveryPricing{
					DistanceRanges: []models.DistanceRange{
						{Min: 0, Max: 500, A: 0, B: 0},
						// Matches nothing > 500m
					},
				},
			},
			wantErr: ErrNoRangeFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculatePrice(tt.input, tt.venue)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("CalculatePrice() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CalculatePrice() unexpected error = %v", err)
			}

			if got.TotalPrice != tt.wantTotal {
				t.Errorf("TotalPrice = %v, want %v", got.TotalPrice, tt.wantTotal)
			}
			if got.Delivery.Fee != tt.wantFee {
				t.Errorf("Fee = %v, want %v", got.Delivery.Fee, tt.wantFee)
			}
			if got.SmallOrderSurcharge != tt.wantSurcharge {
				t.Errorf("Surcharge = %v, want %v", got.SmallOrderSurcharge, tt.wantSurcharge)
			}
		})
	}
}

// Helper to create models.VenueStatic
func newStaticVenueWithLocation(lat, lon float64) models.VenueStatic {
	var v models.VenueStatic
	v.VenueRaw.Location.Coordinates = []float64{lon, lat}
	return v
}

// Helper to create empty coordinates models.VenueStatic
func newStaticVenueWithoutCoords() models.VenueStatic {
	var v models.VenueStatic
	v.VenueRaw.Location.Coordinates = []float64{}
	return v
}

// Helper to create single coordinate models.VenueStatic
func newStaticVenueWithSingleCoord(coordinate float64) models.VenueStatic {
	var v models.VenueStatic
	v.VenueRaw.Location.Coordinates = []float64{coordinate}
	return v
}

// Helper to create models.VenueDynamic with basic pricing
func newDynamicVenueWithPricing(minOrder int, basePrice int) models.VenueDynamic {
	var v models.VenueDynamic
	v.VenueRaw.DeliverySpecs.OrderMinimumNoSurcharge = minOrder
	v.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice = basePrice
	return v
}

func TestMergeToVenueInfo(t *testing.T) {
	tests := []struct {
		name    string
		static  models.VenueStatic
		dynamic models.VenueDynamic
		wantErr error
	}{
		{
			name:    "Success: Valid data",
			static:  newStaticVenueWithLocation(20.0, 10.0),
			dynamic: newDynamicVenueWithPricing(1000, 100),
			wantErr: nil,
		},
		{
			name:    "Failure: Missing coordinates",
			static:  newStaticVenueWithoutCoords(),
			dynamic: models.VenueDynamic{},
			wantErr: ErrInvalidVenueData,
		},
		{
			name:    "Failure: Only one coordinate",
			static:  newStaticVenueWithSingleCoord(10.0),
			dynamic: models.VenueDynamic{},
			wantErr: ErrInvalidVenueData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergeToVenueInfo(tt.static, tt.dynamic)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("MergeToVenueInfo() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("MergeToVenueInfo() unexpected error = %v", err)
			}

			if tt.wantErr == nil {
				if got.Coordinates[1] != 20.0 {
					t.Errorf("Lat mapped incorrectly, got %v", got.Coordinates[1])
				}
			}
		})
	}
}
