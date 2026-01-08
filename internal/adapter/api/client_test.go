package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lnemenl/wolt_1/internal/adapter/api"
	"github.com/lnemenl/wolt_1/internal/core/domain"
)

func TestFetchVenueLocation(t *testing.T) {
	// Aranging
	// Creating a server. The function inside decides what the server says back
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Safety Check: Did our client ask for the right URL?
		// We expect the client to ask for: /home-assignment-api/v1/venues/.../static
		expectedPath := "/home-assignment-api/v1/venues/test-venue/static"

		if r.URL.Path != expectedPath {
			t.Errorf("Client asked for wrong path. Expected %s, got %s", expectedPath, r.URL.Path)
		}

		// Return 200 OK
		w.WriteHeader(http.StatusOK)

		// Return the JSON
		w.Write([]byte(`{"venue_raw": {"location": {"coordinates": [24.93, 60.17]}}}`))
	}))
	// Close the server when the test finishes, or it keeps running
	defer mockServer.Close()

	location, err := api.FetchVenueLocation(mockServer.URL, "test-venue")

	if err != nil {
		t.Fatalf("Expected success, but got error: %v", err)
	}

	expectedLocation := domain.Location{
		Lat: 60.17,
		Lon: 24.93,
	}

	if location != expectedLocation {
		t.Errorf("Data mismatch. Expected %+v, got %+v", expectedLocation, location)
	}
}

func TestFetchDeliveryData(t *testing.T) {
	//Arranging
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/home-assignment-api/v1/venues/test-venue/dynamic"
		if r.URL.Path != expectedPath {
			t.Errorf("Client asked for wrong path. Expected %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)

		w.Write([]byte(`{
            "venue_raw": {
                "delivery_specs": {
                    "delivery_pricing": {
                        "distance_ranges": [
                            {
                                "min": 0,
                                "max": 500,
                                "a": 100,
                                "b": 0,
                                "flag": null
                            },
                            {
                                "min": 500,
                                "max": 0,
                                "a": 200,
                                "b": 0,
                                "flag": null
                            }
                        ]
                    }
                }
            }
        }`))
	}))
	defer mockServer.Close()

	// ACT
	rules, err := api.FetchDeliveryData(mockServer.URL, "test-venue")

	//ASSERT

	if err != nil {
		t.Fatalf("Expected success, got error %v", err)
	}

	if len(rules.PricingRules) != 2 {
		t.Fatalf("Expected 2 rules, got %d", len(rules.PricingRules))
	}

	if rules.PricingRules[0].A != 100 {
		t.Errorf("Expected first rule base price 100, got %d", rules.PricingRules[0].A)
	}
}
