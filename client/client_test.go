package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchVenueData(t *testing.T) {
	t.Log("--- STARTING TEST: TestFetchVenueData ---")

	// 1. START FAKE SERVER
	// This function handles requests from the client.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("[SERVER] Received Request: Method=%s, Path=%s", r.Method, r.URL.Path)

		switch r.URL.Path {
		case "/test-venue/static":
			t.Log("[SERVER] Matched /static endpoint. Sending 200 OK + JSON...")

			// Step 1: Set Status Code
			w.WriteHeader(http.StatusOK)

			// Step 2: Write Body (JSON)
			// Note: We use raw string literals (backticks `) for multi-line strings
			responseBody := `{
				"venue_raw": {
					"location": {
						"coordinates": [24.93, 60.17]
					}
				}
			}`
			w.Write([]byte(responseBody))
			t.Log("[SERVER] Static response sent.")

		case "/test-venue/dynamic":
			t.Log("[SERVER] Matched /dynamic endpoint. Sending 200 OK + JSON...")

			w.WriteHeader(http.StatusOK)
			responseBody := `{
				"venue_raw": {
					"delivery_specs": {
						"order_minimum_no_surcharge": 1000,
						"delivery_pricing": {
							"base_price": 190,
							"distance_ranges": []
						}
					}
				}
			}`
			w.Write([]byte(responseBody))
			t.Log("[SERVER] Dynamic response sent.")

		default:
			t.Logf("[SERVER] ERROR: Unknown path requested: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()
	t.Logf("[TEST] Fake Server is running at: %s", mockServer.URL)

	// 2. CONFIGURE CLIENT
	api := New()
	api.BaseURL = mockServer.URL + "/" // Pointing client to localhost
	t.Logf("[TEST] Configured Client BaseURL: %s", api.BaseURL)

	// 3. EXECUTE
	// We pass "test-venue".
	// The client will build: BaseURL + "test-venue" + "/static"
	t.Log("[TEST] Calling FetchVenueData('test-venue')...")

	static, dynamic, err := api.FetchVenueData("test-venue")

	// 4. ASSERT
	t.Log("[TEST] Function returned. Checking results...")

	if err != nil {
		t.Fatalf("[TEST] FAILED: Expected success, got error: %v", err)
	}

	// Check Static
	t.Logf("[TEST] Static Coords Received: %v", static.VenueRaw.Location.Coordinates)
	if len(static.VenueRaw.Location.Coordinates) != 2 {
		t.Errorf("Expected 2 coordinates, got %d", len(static.VenueRaw.Location.Coordinates))
	}

	// Check Dynamic
	t.Logf("[TEST] Base Price Received: %d", dynamic.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice)
	if dynamic.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice != 190 {
		t.Errorf("Expected base price 190, got %d", dynamic.VenueRaw.DeliverySpecs.DeliveryPricing.BasePrice)
	}

	t.Log("--- FINISHED TEST: TestFetchVenueData ---")
}

func TestFetchVenueData_NetworkError(t *testing.T) {
	t.Log("--- STARTING TEST: NetworkError ---")

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("[SERVER] Received Request: %s. Forcing 500 Error.", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError) // 500
	}))
	defer mockServer.Close()

	api := New()
	api.BaseURL = mockServer.URL + "/"

	t.Log("[TEST] Calling FetchVenueData...")
	// We use _, _, err because we ONLY care about the error in this test
	_, _, err := api.FetchVenueData("test-venue")

	if err == nil {
		t.Fatal("[TEST] FAILED: Expected an error, but got success!")
	}

	t.Logf("[TEST] SUCCESS: Got expected error: %v", err)
	t.Log("--- FINISHED TEST: NetworkError ---")
}

func TestFetchVenueData_DynamicFailure(t *testing.T) {
	t.Log("--- STARTING TEST: DynamicFailure ---")

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test-venue/static":
			// Static works fine!
			t.Log("[SERVER] Static request. Sending 200 OK.")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"venue_raw": {"location": {"coordinates": [0,0]}}}`))

		case "/test-venue/dynamic":
			// Dynamic FAILS!
			t.Log("[SERVER] Dynamic request. Forcing 404 Not Found.")
			w.WriteHeader(http.StatusNotFound) // 404
		}
	}))
	defer mockServer.Close()

	api := New()
	api.BaseURL = mockServer.URL + "/"

	t.Log("[TEST] Calling FetchVenueData...")
	_, _, err := api.FetchVenueData("test-venue")

	// We expect an error because Dynamic failed
	if err == nil {
		t.Fatal("[TEST] FAILED: Expected error (due to dynamic fail), but got success")
	}

	t.Logf("[TEST] SUCCESS: Got expected error: %v", err)
	t.Log("--- FINISHED TEST: DynamicFailure ---")
}
