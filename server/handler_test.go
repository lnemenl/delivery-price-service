package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lnemenl/wolt_1/client"
	"github.com/lnemenl/wolt_1/models"
)

func TestHandleRequest(t *testing.T) {

	// =========================================================================
	// 1. SETUP: The Fake Internet
	// =========================================================================
	mockWoltAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// SCENARIO: SERVER ERROR
		// If the test asks for a "broken-venue", we simulate a crash at Wolt
		if strings.Contains(r.URL.Path, "broken-venue") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// SCENARIO: STATIC DATA
		if strings.Contains(r.URL.Path, "/static") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"venue_raw": {"location": {"coordinates": [24.93, 60.17]}}}`))
			return
		}

		// SCENARIO: DYNAMIC DATA
		if strings.Contains(r.URL.Path, "/dynamic") {
			w.WriteHeader(http.StatusOK)
			// Max distance 1000m
			w.Write([]byte(`{"venue_raw": {"delivery_specs": {"order_minimum_no_surcharge": 1000, "delivery_pricing": {"base_price":190, "distance_ranges": [{"min":0, "max":1000, "a":0, "b":0}]}}}}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockWoltAPI.Close()

	// Setup Handler
	apiClient := client.New()
	apiClient.BaseURL = mockWoltAPI.URL + "/"
	handler := New(apiClient)

	// =========================================================================
	// TEST CASES
	// =========================================================================

	t.Run("1. Happy Path: Valid Request", func(t *testing.T) {
		// Lat 60.17 = Helsinki (Same as venue) -> Distance 0
		url := "/api/v1/delivery-order-price?venue_slug=test-venue&cart_value=1000&user_lat=60.17&user_lon=24.93"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		// Verify JSON
		var resp models.PriceResponse
		json.NewDecoder(rr.Body).Decode(&resp)
		if resp.TotalPrice != 1190 {
			t.Errorf("Expected 1190, got %d", resp.TotalPrice)
		}
	})

	t.Run("2. Validation: Missing Parameters", func(t *testing.T) {
		// Missing 'venue_slug'
		url := "/api/v1/delivery-order-price?cart_value=1000"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rr.Code)
		}
	})

	t.Run("3. Logic Error: Distance Too Long", func(t *testing.T) {
		// Lat 0.0 is far from Helsinki (Lat 60.17). Distance > 1000m
		// Calculator should return error. Handler should map that to 400
		url := "/api/v1/delivery-order-price?venue_slug=test-venue&cart_value=1000&user_lat=0.0&user_lon=0.0"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 (Too Far), got %d", rr.Code)
		}
		// Optional: Check error message
		if !strings.Contains(rr.Body.String(), "no matching") {
			t.Errorf("Expected 'too long' error, got: %s", rr.Body.String())
		}
	})

	t.Run("4. System Error: Wolt API Down", func(t *testing.T) {
		// Use "broken-venue" to trigger the 500 in our mock server
		url := "/api/v1/delivery-order-price?venue_slug=broken-venue&cart_value=1000&user_lat=60.17&user_lon=24.93"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		// Handler should return 500 Internal Server Error
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500 Internal Error, got %d", rr.Code)
		}
	})

	t.Run("5. Method Check: POST Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/delivery-order-price", nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected 405, got %d", rr.Code)
		}
	})

	t.Run("Validation: Negative Cart Value", func(t *testing.T) {
		// Scenario: User sends -100 as money. Impossible in real life.
		url := "/api/v1/delivery-order-price?venue_slug=test&cart_value=-100&user_lat=60&user_lon=24"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		// Expect 400 Bad Request
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		// Expect specific error message
		if !strings.Contains(rr.Body.String(), "cannot be negative") {
			t.Errorf("Expected negative error, got: %s", rr.Body.String())
		}
	})

	t.Run("Validation: Latitude Too High", func(t *testing.T) {
		// Scenario: User sends Lat 91.0 (North of North Pole).
		url := "/api/v1/delivery-order-price?venue_slug=test&cart_value=100&user_lat=91.0&user_lon=24"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "between -90 and 90") {
			t.Errorf("Expected latitude error, got: %s", rr.Body.String())
		}
	})

	t.Run("Validation: Longitude Too Low", func(t *testing.T) {
		// Scenario: User sends Lon -181 (West of the Date Line).
		url := "/api/v1/delivery-order-price?venue_slug=test&cart_value=100&user_lat=60&user_lon=-181"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "between -180 and 180") {
			t.Errorf("Expected longitude error, got: %s", rr.Body.String())
		}
	})

	t.Run("Validation: Garbage Cart Value", func(t *testing.T) {
		// Scenario: User sends text "abc" instead of numbers.
		url := "/api/v1/delivery-order-price?venue_slug=test&cart_value=abc&user_lat=60&user_lon=24"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "invalid cart_value") {
			t.Errorf("Expected invalid format error, got: %s", rr.Body.String())
		}
	})

	t.Run("Validation: Venue Slug Too Long", func(t *testing.T) {
		// Scenario: User sends a slug that is 101 characters long.
		// We use strings.Repeat("a", 101) to generate it easily.
		longSlug := strings.Repeat("a", 101)
		url := "/api/v1/delivery-order-price?venue_slug=" + longSlug + "&cart_value=100&user_lat=60&user_lon=24"

		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "venue_slug too long") {
			t.Errorf("Expected slug length error, got: %s", rr.Body.String())
		}
	})

	t.Run("System: Upstream 404 (Venue Not Found)", func(t *testing.T) {
		// Scenario: User asks for "ghost-venue".
		// Our mock server defaults to 404 for unknown paths, so this simulates a missing venue.
		url := "/api/v1/delivery-order-price?venue_slug=ghost-venue&cart_value=1000&user_lat=60.17&user_lon=24.93"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		// Expect 404 Not Found (NOT 500)
		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected 404 Not Found, got %d", rr.Code)
		}
	})
}
