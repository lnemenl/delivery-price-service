package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchVenueData(t *testing.T) {

	// Set up mock API server to simulate Wolt API
	// This allows testing without external network dependencies
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Handle requests for static venue data
		if strings.Contains(r.URL.Path, "/test-venue/static") {
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
		if strings.Contains(r.URL.Path, "/test-venue/dynamic") {
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
		api := New(mockServer.URL+"/", 10*time.Second)
		// Override base URL to point to local mock server
		api.BaseURL = mockServer.URL + "/"

		static, dynamic, err := api.FetchVenueData(context.Background(), "test-venue")

		if err != nil {
			t.Fatalf("Expected success, but got error: %v", err)
		}

		if len(static.VenueRaw.Location.Coordinates) != 2 {
			t.Errorf("Expected 2 coordinates, got %d", len(static.VenueRaw.Location.Coordinates))
		}
		if static.VenueRaw.Location.Coordinates[0] != 24.93 {
			t.Errorf("Expected lon 24.93, got %f", static.VenueRaw.Location.Coordinates[0])
		}

		if dynamic.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice != 190 {
			t.Errorf("Expected base price 190, got %d", dynamic.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice)
		}
	})

	// Test: Proper error handling for non-existent venue
	// Mock server returns 404 for unknown venue
	t.Run("Sad Path: Dynamic Endpoint Fails", func(t *testing.T) {
		api := New(mockServer.URL+"/", 10*time.Second)
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

func TestFetchVenueData_Cancellation(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Scenario 1: The "Static" endpoint fails immediately.
		if strings.Contains(r.URL.Path, "/static") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Scenario 2: The "Dynamic" endpoint is slow.
		if strings.Contains(r.URL.Path, "/dynamic") {
			// Wait for one of two things to happen:
			select {
			// 1: The "Slow" Path
			case <-time.After(2 * time.Second):
				w.WriteHeader(http.StatusOK)
				// 2: The "Cancel" path
			case <-r.Context().Done():
				return
			}
			return
		}
	}))
	defer mockServer.Close()

	client := New(mockServer.URL+"/", 10*time.Second)
	client.BaseURL = mockServer.URL + "/"

	// Start a stopwatch
	start := time.Now()
	_, _, err := client.FetchVenueData(context.Background(), "test-venue")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected an error (500), but got success")
	}

	// If it took < 100ms, it means successfully cancelled request
	if elapsed > 100*time.Millisecond {
		t.Errorf("Fail: Test took %v. Cancellation is broken.", elapsed)
	} else {
		t.Logf("Success: Test took %v. Cancellation worked!", elapsed)
	}
}
