package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lnemenl/delivery-price-service/client"
	"github.com/lnemenl/delivery-price-service/models"
)

func TestHandleRequest(t *testing.T) {

	// Set up mock API server for testing
	mockWoltAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Return internal server error for broken-venue to simulate API failure
		if strings.Contains(r.URL.Path, "broken-venue") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Return static venue data for test-venue
		if strings.Contains(r.URL.Path, "test-venue") && strings.Contains(r.URL.Path, "/static") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"venue_raw": {"location": {"coordinates": [24.93, 60.17]}}}`))
			return
		}

		// Return dynamic pricing data for test-venue
		if strings.Contains(r.URL.Path, "test-venue") && strings.Contains(r.URL.Path, "/dynamic") {
			w.WriteHeader(http.StatusOK)
			// Max delivery distance set to 1000m. We add the closing range {min:1000, max:0} to simulate "too far".
			response := `{
				"venue_raw": {
					"delivery_specs": {
						"order_minimum_no_surcharge": 1000,
						"delivery_pricing": {
							"base_price": 190,
							"distance_ranges": [
								{"min": 0, "max": 1000, "a": 0, "b": 0},
								{"min": 1000, "max": 0, "a": 0, "b": 0}
							]
						}
					}
				}
			}`
			w.Write([]byte(response))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockWoltAPI.Close()

	// Create handler instance with mock API client
	apiClient := client.New(mockWoltAPI.URL+"/", 10*time.Second)
	apiClient.BaseURL = mockWoltAPI.URL + "/"
	handler := NewDeliveryHandler(apiClient)

	// Run test cases

	t.Run("1. Happy Path: Valid Request", func(t *testing.T) {
		// User location matches venue location (Helsinki), distance = 0
		url := "/api/v1/delivery-order-price?venue_slug=test-venue&cart_value=1000&user_lat=60.17&user_lon=24.93"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		// Verify response JSON is correct
		var resp models.PriceResponse
		json.NewDecoder(rr.Body).Decode(&resp)
		if resp.TotalPrice != 1190 {
			t.Errorf("Expected 1190, got %d", resp.TotalPrice)
		}
	})

	t.Run("2. Validation: Missing Parameters", func(t *testing.T) {
		// Request without venue_slug parameter
		url := "/api/v1/delivery-order-price?cart_value=1000"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rr.Code)
		}
	})

	t.Run("3. Logic Error: Distance Too Long", func(t *testing.T) {
		// User location far from venue (distance > 1000m), delivery not available
		url := "/api/v1/delivery-order-price?venue_slug=test-venue&cart_value=1000&user_lat=0.0&user_lon=0.0"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 (Too Far), got %d", rr.Code)
		}
		// Verify error message indicates distance issue
		if !strings.Contains(rr.Body.String(), "delivery distance too long") {
			t.Errorf("Expected 'too long' error, got: %s", rr.Body.String())
		}
	})

	t.Run("4. System Error: Wolt API Down", func(t *testing.T) {
		// Request broken-venue to trigger server error response
		url := "/api/v1/delivery-order-price?venue_slug=broken-venue&cart_value=1000&user_lat=60.17&user_lon=24.93"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		// Verify handler returns 500 Internal Server Error
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
		// Test rejection of negative cart value
		url := "/api/v1/delivery-order-price?venue_slug=test&cart_value=-100&user_lat=60&user_lon=24"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		// Verify rejection with 400 Bad Request
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		// Verify error message contains expected text
		if !strings.Contains(rr.Body.String(), "cannot be negative") {
			t.Errorf("Expected negative error, got: %s", rr.Body.String())
		}
	})

	t.Run("Validation: Latitude Too High", func(t *testing.T) {
		// Test rejection of invalid latitude (91.0 > 90)
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
		// Test rejection of invalid longitude (-181 < -180)
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
		// Test rejection of non-numeric cart value
		url := "/api/v1/delivery-order-price?venue_slug=test&cart_value=abc&user_lat=60&user_lon=24"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "invalid cart_value") {
			t.Errorf("Expected invalid cart_value error, got: %s", rr.Body.String())
		}
	})

	t.Run("Validation: Venue Slug Too Long", func(t *testing.T) {
		// Test rejection of slug exceeding maximum length
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
		// Request non-existent venue to test 404 handling
		url := "/api/v1/delivery-order-price?venue_slug=ghost-venue&cart_value=1000&user_lat=60.17&user_lon=24.93"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		// Verify 404 Not Found response for missing venue
		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected 404 Not Found, got %d", rr.Code)
		}
	})

	t.Run("Validation: Missing cart_value", func(t *testing.T) {
		url := "/api/v1/delivery-order-price?venue_slug=test&user_lat=60&user_lon=24"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "missing cart_value") {
			t.Errorf("Expected missing cart_value error, got: %s", rr.Body.String())
		}
	})

	t.Run("Validation: Missing user_lat", func(t *testing.T) {
		url := "/api/v1/delivery-order-price?venue_slug=test&cart_value=100&user_lon=24"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "missing user_lat") {
			t.Errorf("Expected missing user_lat error, got: %s", rr.Body.String())
		}
	})

	t.Run("Validation: Missing user_lon", func(t *testing.T) {
		url := "/api/v1/delivery-order-price?venue_slug=test&cart_value=100&user_lat=60"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handler.HandleRequest(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "missing user_lon") {
			t.Errorf("Expected missing user_lon error, got: %s", rr.Body.String())
		}
	})
}
