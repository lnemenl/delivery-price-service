package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchVenueData(t *testing.T) {

	// Set up mock API server to simulate Wolt API
	// This allows testing without external network dependencies
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Handle requests for static venue data
		if r.URL.Path == "/test-venue/static" {
			w.WriteHeader(http.StatusOK)
			// Return venue location coordinates
			w.Write([]byte(`{
				"venue_raw": {
					"location": {
						"coordinates": [24.93, 60.17]
					}
				}
			}`))
			return
		}

		// Handle requests for dynamic pricing data
		if r.URL.Path == "/test-venue/dynamic" {
			w.WriteHeader(http.StatusOK)
			// Return pricing and delivery specifications
			w.Write([]byte(`{
				"venue_raw": {
					"delivery_specs": {
						"order_minimum_no_surcharge": 1000,
						"delivery_pricing": {
							"base_price": 190,
							"distance_ranges": []
						}
					}
				}
			}`))
			return
		}

		// Return 404 for unknown routes
		w.WriteHeader(http.StatusNotFound)
	}))
	// Clean up mock server after test
	defer mockServer.Close()

	// Test: Successful fetch of both static and dynamic data
	// Mock server returns valid JSON for test-venue
	t.Run("Happy Path: Success Fetch", func(t *testing.T) {
		// Create API client instance
		api := New()
		// Override base URL to point to local mock server
		api.BaseURL = mockServer.URL + "/"

		// Call FetchVenueData
		static, dynamic, err := api.FetchVenueData(context.Background(), "test-venue")

		// Verify both static and dynamic data are correctly decoded
		if err != nil {
			t.Fatalf("Expected success, but got error: %v", err)
		}

		// Verify static data coordinates
		if len(static.VenueRaw.Location.Coordinates) != 2 {
			t.Errorf("Expected 2 coordinates, got %d", len(static.VenueRaw.Location.Coordinates))
		}
		if static.VenueRaw.Location.Coordinates[0] != 24.93 {
			t.Errorf("Expected lon 24.93, got %f", static.VenueRaw.Location.Coordinates[0])
		}

		// Verify dynamic data pricing
		if dynamic.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice != 190 {
			t.Errorf("Expected base price 190, got %d", dynamic.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice)
		}
	})

	// Test: Proper error handling for non-existent venue
	// Mock server returns 404 for unknown venue
	t.Run("Sad Path: Dynamic Endpoint Fails", func(t *testing.T) {
		api := New()
		api.BaseURL = mockServer.URL + "/"

		// Request non-existent venue
		_, _, err := api.FetchVenueData(context.Background(), "wrong-venue")

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected an error (404), but got success")
		}

		// Verify error type using errors.Is
		if !errors.Is(err, ErrVenueNotFound) {
			t.Errorf("Expected error to wrap ErrVenueNotFound, got: %v", err)
		}

		// Verify context wrapping
		actualError := err.Error()
		if !strings.Contains(actualError, "static data error") && !strings.Contains(actualError, "dynamic data error") {
			t.Errorf("Expected error to be about static or dynamic data, got %q", actualError)
		}
	})
}
