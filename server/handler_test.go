package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lnemenl/wolt_1/client"
	"github.com/lnemenl/wolt_1/models"
)

func TestHandleRequest(t *testing.T) {
	t.Log("--- STARTING TEST: TestHandleRequest ---")

	// ---------------------------------------------------------
	// 1. SETUP THE FAKE INTERNET (Mock Wolt API)
	// ---------------------------------------------------------
	// The Handler needs a Client. The Client needs an API.
	// We cannot use the real Wolt API, so we build a fake one here.
	mockWoltAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("[FAKE WOLT API] Received request: %s", r.URL.Path)

		// A. Static Data Endpoint
		if r.URL.Path == "/test-venue/static" {
			t.Log("[FAKE WOLT API] Serving Static Data...")
			w.WriteHeader(http.StatusOK)
			// Return coordinates [24.93, 60.17]
			w.Write([]byte(`{"venue_raw": {"location": {"coordinates": [24.93, 60.17]}}}`))
			return
		}

		// B. Dynamic Data Endpoint
		if r.URL.Path == "/test-venue/dynamic" {
			t.Log("[FAKE WOLT API] Serving Dynamic Data...")
			w.WriteHeader(http.StatusOK)
			// Return Base Price: 190, Min Order: 1000
			w.Write([]byte(`{
				"venue_raw": {
					"delivery_specs": {
						"order_minimum_no_surcharge": 1000,
						"delivery_pricing": {
							"base_price": 190,
							"distance_ranges": [{"min": 0, "max": 0, "a": 0, "b": 0}]
						}
					}
				}
			}`))
			return
		}

		t.Logf("[FAKE WOLT API] Error: Unknown path %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	// Ensure we shut down the fake internet when test ends
	defer mockWoltAPI.Close()
	t.Logf("[TEST] Fake Internet running at: %s", mockWoltAPI.URL)

	// ---------------------------------------------------------
	// 2. CONSTRUCT THE HANDLER CHAIN
	// ---------------------------------------------------------

	// A. Create Client (The Tool)
	apiClient := client.New()
	// IMPORTANT: Point the tool to our Fake Internet, not the real one
	apiClient.BaseURL = mockWoltAPI.URL + "/"
	t.Log("[TEST] Client created and pointed to Fake Internet.")

	// B. Create Handler (The Worker)
	// We inject the client so the handler can use it.
	handler := New(apiClient)
	t.Log("[TEST] Handler created with injected Client.")

	// ---------------------------------------------------------
	// 3. SIMULATE THE BROWSER (The Request)
	// ---------------------------------------------------------
	// We manually build the URL string exactly as a user would type it.
	// slug: "test-venue"
	// cart: 1000
	// lat/lon: Same as venue (distance = 0)
	targetURL := "/api/v1/delivery-order-price?venue_slug=test-venue&cart_value=1000&user_lat=60.17&user_lon=24.93"

	t.Logf("[TEST] Building Fake Request: GET %s", targetURL)
	req := httptest.NewRequest(http.MethodGet, targetURL, nil)

	// ---------------------------------------------------------
	// 4. PREPARE THE RECORDER (The Screen)
	// ---------------------------------------------------------
	// 'rr' = Response Recorder. It acts like a browser tab waiting for data.
	rr := httptest.NewRecorder()
	t.Log("[TEST] Recorder ready to catch response.")

	// ---------------------------------------------------------
	// 5. ACTION!
	// ---------------------------------------------------------
	t.Log("[TEST] Calling handler.HandleRequest(rr, req)...")

	// We pass the Input (req) and the Output Pipe (rr) to the handler.
	handler.HandleRequest(rr, req)

	t.Log("[TEST] Handler finished. Checking Recorder contents...")

	// ---------------------------------------------------------
	// 6. INSPECT RESULTS
	// ---------------------------------------------------------

	// A. Check Status Code (The Envelope Stamp)
	t.Logf("[TEST] Status Code Received: %d", rr.Code)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}

	// B. Check Body (The JSON Content)
	// rr.Body holds the bytes written by the handler.
	t.Logf("[TEST] Raw Body Received: %s", rr.Body.String())

	var response models.PriceResponse
	// We decode the JSON from the recorder into a struct so we can check numbers
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	// C. Validate Math
	// Cart Value: 1000
	// Delivery Fee: 190 (Base) + 0 (Distance) = 190
	// Surcharge: 0 (Cart 1000 >= Min 1000)
	// Total Expected: 1190

	t.Logf("[TEST] Calculated Total Price: %d", response.TotalPrice)

	if response.TotalPrice != 1190 {
		t.Errorf("Math mismatch! Expected 1190, got %d", response.TotalPrice)
	} else {
		t.Log("[TEST] Math is correct!")
	}

	t.Log("--- FINISHED TEST: TestHandleRequest ---")
}

func TestHandleRequest_EdgeCases(t *testing.T) {
	t.Log("--- STARTING TEST: Edge Cases ---")

	// 1. SETUP FAKE INTERNET
	// This time, we make it simple: it always returns 200 OK.
	// We are testing OUR validation logic, not the API client again.
	mockWoltAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return valid minimal JSON so the client doesn't crash during JSON decoding
		w.Write([]byte(`{"venue_raw": {"location": {"coordinates": [0,0]}, "delivery_specs": {"order_minimum_no_surcharge": 0, "delivery_pricing": {"base_price":0, "distance_ranges": []}}}}`))
	}))
	defer mockWoltAPI.Close()

	// 2. SETUP HANDLER
	apiClient := client.New()
	apiClient.BaseURL = mockWoltAPI.URL + "/"
	handler := New(apiClient)

	// 3. DEFINE SCENARIOS
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "Wrong Method (POST)",
			method:         http.MethodPost, // <--- POST instead of GET
			url:            "/api/v1/price?venue_slug=test&cart_value=100",
			expectedStatus: http.StatusMethodNotAllowed, // 405
		},
		{
			name:           "Missing Venue Slug",
			method:         http.MethodGet,
			url:            "/api/v1/price?cart_value=100", // <--- No venue_slug
			expectedStatus: http.StatusBadRequest,          // 400
		},
		{
			name:           "Invalid Cart Value (Letters)",
			method:         http.MethodGet,
			url:            "/api/v1/price?venue_slug=test&cart_value=abc", // <--- "abc" is not int
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Coordinate (Letters)",
			method:         http.MethodGet,
			url:            "/api/v1/price?venue_slug=test&cart_value=100&user_lat=oops", // <--- "oops" is not float
			expectedStatus: http.StatusBadRequest,
		},
	}

	// 4. RUN THE LOOP
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Request
			req := httptest.NewRequest(tt.method, tt.url, nil)

			// Create Recorder
			rr := httptest.NewRecorder()

			// Action
			handler.HandleRequest(rr, req)

			// Assert
			if rr.Code != tt.expectedStatus {
				t.Errorf("Scenario '%s' failed! Expected status %d, got %d",
					tt.name, tt.expectedStatus, rr.Code)
			} else {
				t.Logf("Scenario '%s': Passed (Got %d)", tt.name, rr.Code)
			}
		})
	}
	t.Log("--- FINISHED TEST: Edge Cases ---")
}
