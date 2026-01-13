package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchVenueData(t *testing.T) {

	// =========================================================================
	// 1. SETTING THE SCENE (The Fake Internet)
	// We create a local web server to mimic Wolt. This allows us to test
	// without actual internet access and ensures consistent results
	// =========================================================================
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// SCENARIO: Client asks for STATIC data
		// We listen for the specific URL ending in "/static"
		if r.URL.Path == "/test-venue/static" {
			w.WriteHeader(http.StatusOK)
			// Return minimal valid JSON for location
			w.Write([]byte(`{
				"venue_raw": {
					"location": {
						"coordinates": [24.93, 60.17]
					}
				}
			}`))
			return
		}

		// SCENARIO: Client asks for DYNAMIC data
		// We listen for the specific URL ending in "/dynamic"
		if r.URL.Path == "/test-venue/dynamic" {
			w.WriteHeader(http.StatusOK)
			// Return minimal valid JSON for pricing
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

		// FALLBACK: If the URL is wrong, return 404 Not Found
		w.WriteHeader(http.StatusNotFound)
	}))
	// Clean up: Shut down the server when the test finishes
	defer mockServer.Close()

	// =========================================================================
	// CHAPTER 1: THE SUCCESSFUL DOWNLOAD
	// Scenario: Everything works. We ask for "test-venue", and the mock server
	// returns valid JSON for both static and dynamic calls
	// =========================================================================
	t.Run("Happy Path: Success Fetch", func(t *testing.T) {
		// 1. Setup the Tool
		api := New()
		// CRITICAL: We override the BaseURL to point to our local mock server
		// instead of the real Wolt API
		api.BaseURL = mockServer.URL + "/"

		// 2. Action: Call the function
		static, dynamic, err := api.FetchVenueData("test-venue")

		// 3. Assertions
		if err != nil {
			t.Fatalf("Expected success, but got error: %v", err)
		}

		// Check if Static data was correctly decoded
		if len(static.VenueRaw.Location.Coordinates) != 2 {
			t.Errorf("Expected 2 coordinates, got %d", len(static.VenueRaw.Location.Coordinates))
		}
		if static.VenueRaw.Location.Coordinates[0] != 24.93 {
			t.Errorf("Expected lon 24.93, got %f", static.VenueRaw.Location.Coordinates[0])
		}

		// Check if Dynamic data was correctly decoded
		if dynamic.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice != 190 {
			t.Errorf("Expected base price 190, got %d", dynamic.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice)
		}
	})

	// =========================================================================
	// CHAPTER 2: THE BROKEN LINK
	// Scenario: We ask for a venue that doesn't exist ("wrong-venue")
	// The mock server will return 404 (Not Found)
	// =========================================================================
	t.Run("Sad Path: Dynamic Endpoint Fails", func(t *testing.T) {
		api := New()
		api.BaseURL = mockServer.URL + "/"

		// We ask for "wrong-venue". Our mock server logic above defaults to 404
		_, _, err := api.FetchVenueData("wrong-venue")

		// We expect an error here.
		if err == nil {
			t.Fatal("Expected an error (404), but got success")
		}

		// Verify the error message is what we expect
		// The client.go adds context "static data error: ..." or "dynamic data error: ..."
		// Since static is called first, it should fail there first
		expectedError := "API returned status: 404"
		if fmt.Sprintf("%s", err) != "static data error: "+expectedError {
			t.Logf("Got expected error: %v", err)
		}
	})
}
